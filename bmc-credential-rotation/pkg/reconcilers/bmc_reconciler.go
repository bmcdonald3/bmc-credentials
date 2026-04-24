package reconcilers

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	v1 "github.com/example/bmc-credential-rotation/apis/example.fabrica.dev/v1"
)

func (r *BmcReconciler) reconcileBmc(ctx context.Context, res *v1.Bmc) error {
	if res.Status.Phase == "Aligned" {
		return nil
	}

	if err := r.updateStatus(ctx, res, "Reconciling", ""); err != nil {
		return fmt.Errorf("failed to set reconciling state: %w", err)
	}

	httpClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	baseURL := fmt.Sprintf("https://%s", res.Spec.TargetIP)
	ifaceURL := baseURL + "/redfish/v1/Managers/BMC/EthernetInterfaces/1"
	acctURL := baseURL + "/redfish/v1/AccountService/Accounts/1"

	ifaceBody, statusCode, responseBody, err := redfishRequest(ctx, httpClient, http.MethodGet, ifaceURL, res.Spec.DesiredAuth, nil)
	if err != nil {
		return r.failReconciliation(ctx, res, err)
	}

	if statusCode == http.StatusUnauthorized {
		authPayload := map[string]interface{}{
			"Password": res.Spec.DesiredAuth.Password,
			"Enabled":  true,
			"RoleId":   "Administrator",
		}

		_, statusCode, responseBody, err = redfishRequest(ctx, httpClient, http.MethodPatch, acctURL, res.Spec.CurrentAuth, authPayload)
		if err != nil {
			return r.failReconciliation(ctx, res, err)
		}
		if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
			return r.failReconciliation(ctx, res, fmt.Errorf("http %d: %s", statusCode, responseBody))
		}

		ifaceBody, statusCode, responseBody, err = redfishRequest(ctx, httpClient, http.MethodGet, ifaceURL, res.Spec.DesiredAuth, nil)
		if err != nil {
			return r.failReconciliation(ctx, res, err)
		}
	}

	if statusCode != http.StatusOK {
		return r.failReconciliation(ctx, res, fmt.Errorf("http %d: %s", statusCode, responseBody))
	}

	alreadyAligned, subnetMask, gateway, err := parseIPv4State(ifaceBody, res.Spec.TargetIP)
	if err != nil {
		return r.failReconciliation(ctx, res, err)
	}
	if alreadyAligned {
		if err := r.updateStatus(ctx, res, "Aligned", ""); err != nil {
			return fmt.Errorf("failed to set aligned state: %w", err)
		}
		return nil
	}

	networkPayload := map[string]interface{}{
		"DHCPv4": map[string]interface{}{
			"DHCPEnabled": false,
		},
		"IPv4StaticAddresses": []map[string]interface{}{
			{
				"Address":    res.Spec.TargetIP,
				"SubnetMask": subnetMask,
				"Gateway":    gateway,
			},
		},
	}

	_, statusCode, responseBody, err = redfishRequest(ctx, httpClient, http.MethodPatch, ifaceURL, res.Spec.DesiredAuth, networkPayload)
	if err != nil {
		return r.failReconciliation(ctx, res, err)
	}
	if statusCode != http.StatusOK && statusCode != http.StatusNoContent {
		return r.failReconciliation(ctx, res, fmt.Errorf("http %d: %s", statusCode, responseBody))
	}

	if err := r.updateStatus(ctx, res, "Aligned", ""); err != nil {
		return fmt.Errorf("failed to set aligned state: %w", err)
	}

	r.Logger.Infof("Successfully reconciled Bmc %s", res.GetUID())
	return nil
}

func (r *BmcReconciler) updateStatus(ctx context.Context, res *v1.Bmc, phase, message string) error {
	now := time.Now()
	res.Status.Phase = phase
	res.Status.Message = message
	res.Status.LastUpdated = &now
	return r.UpdateStatus(ctx, res)
}

func (r *BmcReconciler) failReconciliation(ctx context.Context, res *v1.Bmc, cause error) error {
	message := strings.TrimSpace(cause.Error())
	if err := r.updateStatus(ctx, res, "Failed", message); err != nil {
		return fmt.Errorf("failed to persist failed state: %w (original error: %v)", err, cause)
	}
	return cause
}

func redfishRequest(ctx context.Context, client *http.Client, method, url string, auth v1.Credentials, payload interface{}) ([]byte, int, string, error) {
	var bodyReader io.Reader
	if payload != nil {
		body, err := json.Marshal(payload)
		if err != nil {
			return nil, 0, "", fmt.Errorf("failed to marshal request payload: %w", err)
		}
		bodyReader = strings.NewReader(string(body))
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, 0, "", fmt.Errorf("failed to build request: %w", err)
	}
	req.SetBasicAuth(auth.Username, auth.Password)
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, "", fmt.Errorf("request %s %s failed: %w", method, url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, "", fmt.Errorf("failed to read response body: %w", err)
	}

	bodyText := strings.TrimSpace(string(respBody))
	return respBody, resp.StatusCode, bodyText, nil
}

func parseIPv4State(raw []byte, targetIP string) (bool, string, string, error) {
	var data struct {
		IPv4Addresses []struct {
			Address    string `json:"Address"`
			SubnetMask string `json:"SubnetMask"`
			Gateway    string `json:"Gateway"`
		} `json:"IPv4Addresses"`
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return false, "", "", fmt.Errorf("failed to parse IPv4Addresses response: %w", err)
	}

	for _, addr := range data.IPv4Addresses {
		if addr.Address == targetIP {
			return true, addr.SubnetMask, addr.Gateway, nil
		}
	}

	for _, addr := range data.IPv4Addresses {
		if addr.SubnetMask != "" && addr.Gateway != "" {
			return false, addr.SubnetMask, addr.Gateway, nil
		}
	}

	return false, "", "", fmt.Errorf("unable to extract SubnetMask and Gateway from IPv4Addresses")
}
