package alert

import (
	"encoding/json"
	"testing"
)

func TestUmarshallPipeline(t *testing.T) {
	a := `{"annotations":{"description":"job has disappeared","summary":"job missing"},"endsAt":"2026-02-03T15:48:53.926Z","fingerprint":"17358d92dd0f3b58","receivers":[{"name":"team-X-mails"}],"startsAt":"2026-02-03T15:12:53.926Z","status":{"inhibitedBy":[],"mutedBy":[],"silencedBy":["d2353af6"],"state":"suppressed"},"updatedAt":"2026-02-03T15:44:53.929Z","generatorURL":"http://f2526c40017b:9090","labels":{"alertname":"missing","job":"alertmanager","monitor":"my-monitor","severity":"low"}}`

	var alert AlertmanagerAlert
	err := json.Unmarshal([]byte(a), &alert)

	if err != nil {
		t.Error(err)
	}

	if alert.Annotations.Summary != "job missing" {
		t.Error("\nActual: ", alert.Annotations.Summary, "\nExpected: ", "job missing")
	}
}
