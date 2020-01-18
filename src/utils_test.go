package databunker

import (
	"net/http/httptest"
	"strings"
	"testing"

	uuid "github.com/hashicorp/go-uuid"
)

func Test_UUID(t *testing.T) {
	for id := 1; id < 11; id++ {
		recordUUID, err := uuid.GenerateUUID()
		t.Logf("Checking[%d]: %s\n", id, recordUUID)
		if err != nil {
			t.Fatalf("Failed to generate UUID %s: %s ", recordUUID, err)
		} else if isValidUUID(recordUUID) == false {
			t.Fatalf("Failed to validate UUID: %s ", recordUUID)
		}
	}
}

func Test_AppNames(t *testing.T) {
	goodApps := []string{"penn", "teller", "a123"}
	for _, value := range goodApps {
		if isValidApp(value) == false {
			t.Fatalf("Failed to validate good app name: %s ", value)
		}
	}
	badApps := []string{"P1", "1as", "_a", "a_a", "a.a", "a a"}
	for _, value := range badApps {
		if isValidApp(value) == true {
			t.Fatalf("Failed to validate bad app name: %s ", value)
		}
	}

}

func Test_getJSONPost(t *testing.T) {
	goodJsons := []string{
		`{"login":"abc","name": "tom", "pass": "mylittlepony", "admin": true}`,
		`{"login":"1234","name": "tom", "pass": "mylittlepony", "admin": true}`,
	}
	for _, value := range goodJsons {
		request := httptest.NewRequest("POST", "/user", strings.NewReader(value))
		request.Header.Set("Content-Type", "application/json")
		result, err := getJSONPost(request, "IL")
		if err != nil {
			t.Fatalf("Failed to parse json: %s, err: %s\n", value, err)
		}
		if len(result.loginIdx) == 0 {
			t.Fatalf("Failed to parse login index from json: %s ", value)
		}
	}

	badJsons := []string{
		`{"login":true,"name": "tom", "pass": "mylittlepony", "admin": true}`,
		`{"login":1,"name": "tom", "pass": "mylittlepony", "admin": true}`,
		`{"login":null,"name": "tom", "pass": "mylittlepony", "admin": true}`,
	}
	for _, value := range badJsons {
		request := httptest.NewRequest("POST", "/user", strings.NewReader(value))
		request.Header.Set("Content-Type", "application/json")
		result, err := getJSONPost(request, "IL")
		if err != nil {
			t.Fatalf("Failed to parse json: %s, err: %s\n", value, err)
		}
		if len(result.loginIdx) != 0 {
			t.Fatalf("Failed to parse login index from json: %s ", value)
		}
	}
}
