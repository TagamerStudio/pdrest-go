package pdrest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	contractRevision = "2026-09-17"
	contractPrefix   = "/v1/pdapi"
	contractToken    = "contract-token"
	contractPlayer   = "steam_76561198012345678"
	contractGuildID  = "f0a1c3e9-7d5b-4a28-8c33-411fdc2e6b74"
	contractBaseID   = "13b9e8d7-4f2c-42a1-b79e-fc2a9186e4d5"
	contractIP       = "203.0.113.7"
)

type contractEndpoint struct {
	name          string
	method        string
	path          string
	requestBody   string
	responseBody  string
	errorStatuses []int
	invoke        func(*Client) error
}

func TestDocumentedRESTContract_Success(t *testing.T) {
	for _, endpoint := range documentedContractEndpoints(t) {
		t.Run(endpoint.name, func(t *testing.T) {
			server := newContractServer(t, endpoint, http.StatusOK)
			defer server.Close()

			client, err := NewClient(server.URL, contractToken)
			if err != nil {
				t.Fatalf("NewClient failed: %v", err)
			}
			if err := endpoint.invoke(client); err != nil {
				t.Fatalf("endpoint failed: %v", err)
			}
		})
	}
}

func TestDocumentedRESTContract_ErrorStatuses(t *testing.T) {
	for _, endpoint := range documentedContractEndpoints(t) {
		for _, status := range endpoint.errorStatuses {
			t.Run(fmt.Sprintf("%s/%d", endpoint.name, status), func(t *testing.T) {
				assertContractError(t, endpoint, status)
			})
		}
	}
}

func assertContractError(t *testing.T, endpoint contractEndpoint, status int) {
	t.Helper()

	server := newContractServer(t, endpoint, status)
	defer server.Close()

	client, err := NewClient(server.URL, contractToken)
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	err = endpoint.invoke(client)
	if err == nil {
		t.Fatalf("expected status %d error", status)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != status {
		t.Fatalf("unexpected status: got %d, want %d", apiErr.StatusCode, status)
	}
	if apiErr.Envelope == nil || apiErr.Envelope.Error.Code != "ERROR_CODE" {
		t.Fatalf("expected documented error envelope, got %+v", apiErr.Envelope)
	}
}

func documentedContractEndpoints(t *testing.T) []contractEndpoint {
	t.Helper()

	postResponses := readContractResponses(t)
	errorStatuses := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusInternalServerError,
	}
	notFoundStatuses := append(append([]int{}, errorStatuses...), http.StatusNotFound)

	return []contractEndpoint{
		{
			name:          "version",
			method:        http.MethodGet,
			path:          "/version",
			responseBody:  readContractFixture(t, "version.json"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetVersion(context.Background())
				return err
			},
		},
		{
			name:          "guilds",
			method:        http.MethodGet,
			path:          "/guilds",
			responseBody:  readContractFixture(t, "guilds.json"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetGuilds(context.Background())
				return err
			},
		},
		{
			name:          "guild",
			method:        http.MethodGet,
			path:          "/guild/" + contractGuildID,
			responseBody:  readContractFixture(t, "guild.json"),
			errorStatuses: notFoundStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetGuild(context.Background(), contractGuildID)
				return err
			},
		},
		{
			name:          "players",
			method:        http.MethodGet,
			path:          "/players",
			responseBody:  readContractFixture(t, "players.json"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetPlayers(context.Background())
				return err
			},
		},
		{
			name:          "player",
			method:        http.MethodGet,
			path:          "/player/" + contractPlayer,
			responseBody:  readContractFixture(t, "player.json"),
			errorStatuses: notFoundStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetPlayer(context.Background(), contractPlayer)
				return err
			},
		},
		{
			name:          "pals",
			method:        http.MethodGet,
			path:          "/pals/" + contractPlayer,
			responseBody:  readContractFixture(t, "pals.json"),
			errorStatuses: notFoundStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetPals(context.Background(), contractPlayer)
				return err
			},
		},
		{
			name:          "items",
			method:        http.MethodGet,
			path:          "/items/" + contractPlayer,
			responseBody:  readContractFixture(t, "items.json"),
			errorStatuses: notFoundStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetItems(context.Background(), contractPlayer)
				return err
			},
		},
		{
			name:          "techs",
			method:        http.MethodGet,
			path:          "/techs/" + contractPlayer,
			responseBody:  readContractFixture(t, "techs.json"),
			errorStatuses: notFoundStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetTechs(context.Background(), contractPlayer)
				return err
			},
		},
		{
			name:          "progression",
			method:        http.MethodGet,
			path:          "/progression/" + contractPlayer,
			responseBody:  readContractFixture(t, "progression.json"),
			errorStatuses: notFoundStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetProgression(context.Background(), contractPlayer)
				return err
			},
		},
		{
			name:          "banlist",
			method:        http.MethodGet,
			path:          "/banlist",
			responseBody:  readContractFixture(t, "banlist.json"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GetBanlist(context.Background(), nil)
				return err
			},
		},
		{
			name:          "give-items",
			method:        http.MethodPost,
			path:          "/give/items/" + contractPlayer,
			requestBody:   `{"Items":[{"ItemID":"Money","Count":1},{"ItemID":"Wood","Count":2}]}`,
			responseBody:  postResponse(t, postResponses, "/give/items/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GiveItems(context.Background(), contractPlayer, "Money", []any{"Wood", 2})
				return err
			},
		},
		{
			name:          "give-pals",
			method:        http.MethodPost,
			path:          "/give/pals/" + contractPlayer,
			requestBody:   `{"Pals":[{"PalID":"Foxparks","Level":12}]}`,
			responseBody:  postResponse(t, postResponses, "/give/pals/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GivePals(context.Background(), contractPlayer, []any{"Foxparks", 12})
				return err
			},
		},
		{
			name:          "give-paltemplate",
			method:        http.MethodPost,
			path:          "/give/paltemplate/" + contractPlayer,
			requestBody:   `{"PalTemplates":["ArenaBoss.json"]}`,
			responseBody:  postResponse(t, postResponses, "/give/paltemplate/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GivePalTemplates(context.Background(), contractPlayer, "ArenaBoss.json")
				return err
			},
		},
		{
			name:          "give-paleggs",
			method:        http.MethodPost,
			path:          "/give/paleggs/" + contractPlayer,
			requestBody:   `{"PalEggs":[{"EggID":"PalEgg_Fire_01","PalID":"Foxparks","Level":12}]}`,
			responseBody:  postResponse(t, postResponses, "/give/paleggs/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.GivePalEggs(context.Background(), contractPlayer, map[string]any{
					"EggID": "PalEgg_Fire_01",
					"PalID": "Foxparks",
					"Level": 12,
				})
				return err
			},
		},
		{
			name:          "give-progression",
			method:        http.MethodPost,
			path:          "/give/progression/" + contractPlayer,
			requestBody:   `{"EXP":1000}`,
			responseBody:  postResponse(t, postResponses, "/give/progression/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				exp := 1000
				_, err := client.GiveProgression(context.Background(), contractPlayer, nil, &exp, nil, nil, nil)
				return err
			},
		},
		{
			name:          "learntech",
			method:        http.MethodPost,
			path:          "/learntech/" + contractPlayer,
			requestBody:   `{"Technology":"Technology_1"}`,
			responseBody:  postResponse(t, postResponses, "/learntech/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.LearnTech(context.Background(), contractPlayer, "Technology_1")
				return err
			},
		},
		{
			name:          "forgettech",
			method:        http.MethodPost,
			path:          "/forgettech/" + contractPlayer,
			requestBody:   `{"Technology":"All"}`,
			responseBody:  postResponse(t, postResponses, "/forgettech/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.ForgetTech(context.Background(), contractPlayer, "All")
				return err
			},
		},
		{
			name:          "deletebase",
			method:        http.MethodPost,
			path:          "/deletebase/" + contractBaseID,
			requestBody:   `{}`,
			responseBody:  postResponse(t, postResponses, "/deletebase/"+contractBaseID),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.DeleteBase(context.Background(), contractBaseID)
				return err
			},
		},
		{
			name:          "summon-pal",
			method:        http.MethodPost,
			path:          "/summon/pal",
			requestBody:   `{"PalID":"Foxparks","X":230,"Y":-486,"Z":4097,"Level":12}`,
			responseBody:  postResponse(t, postResponses, "/summon/pal"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.SummonPal(context.Background(), &SummonPalRequest{
					PalID: "Foxparks",
					X:     230,
					Y:     -486,
					Z:     4097,
					Level: 12,
				})
				return err
			},
		},
		{
			name:          "summon-npc",
			method:        http.MethodPost,
			path:          "/summon/npc",
			requestBody:   `{"NPCID":"PIDF_Soldier_AssaultRifle","X":230,"Y":-486,"Z":4097}`,
			responseBody:  postResponse(t, postResponses, "/summon/npc"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.SummonNPC(context.Background(), &SummonNPCRequest{
					NPCID: "PIDF_Soldier_AssaultRifle",
					X:     230,
					Y:     -486,
					Z:     4097,
				})
				return err
			},
		},
		{
			name:          "ban",
			method:        http.MethodPost,
			path:          "/ban/" + contractPlayer,
			requestBody:   `{"Reason":"Rule violation"}`,
			responseBody:  postResponse(t, postResponses, "/ban/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.Ban(context.Background(), contractPlayer, "Rule violation", false)
				return err
			},
		},
		{
			name:          "unban",
			method:        http.MethodPost,
			path:          "/unban/" + contractPlayer,
			requestBody:   `{"Reason":"Appeal accepted"}`,
			responseBody:  postResponse(t, postResponses, "/unban/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.Unban(context.Background(), contractPlayer, "Appeal accepted")
				return err
			},
		},
		{
			name:          "banip",
			method:        http.MethodPost,
			path:          "/banip/" + contractIP,
			requestBody:   `{"Reason":"Rule violation"}`,
			responseBody:  postResponse(t, postResponses, "/banip/"+contractIP),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.BanIP(context.Background(), contractIP, &BanIPRequest{Reason: "Rule violation"})
				return err
			},
		},
		{
			name:          "unbanip",
			method:        http.MethodPost,
			path:          "/unbanip/" + contractIP,
			requestBody:   `{"Reason":"Appeal accepted"}`,
			responseBody:  postResponse(t, postResponses, "/unbanip/"+contractIP),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.UnbanIP(context.Background(), contractIP, &UnbanIPRequest{Reason: "Appeal accepted"})
				return err
			},
		},
		{
			name:          "kick",
			method:        http.MethodPost,
			path:          "/kick/" + contractPlayer,
			requestBody:   `{"Reason":"Please reconnect"}`,
			responseBody:  postResponse(t, postResponses, "/kick/"+contractPlayer),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.Kick(context.Background(), contractPlayer, "Please reconnect")
				return err
			},
		},
		{
			name:          "broadcast",
			method:        http.MethodPost,
			path:          "/Broadcast",
			requestBody:   `{"Message":"Restart in 15 minutes."}`,
			responseBody:  postResponse(t, postResponses, "/Broadcast"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.Broadcast(context.Background(), "Restart in 15 minutes.", "")
				return err
			},
		},
		{
			name:          "alert",
			method:        http.MethodPost,
			path:          "/Alert",
			requestBody:   `{"Message":"Restart now."}`,
			responseBody:  postResponse(t, postResponses, "/Alert"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.Alert(context.Background(), "Restart now.")
				return err
			},
		},
		{
			name:          "reloadconfig",
			method:        http.MethodPost,
			path:          "/ReloadConfig",
			requestBody:   `{}`,
			responseBody:  postResponse(t, postResponses, "/ReloadConfig"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.ReloadConfig(context.Background())
				return err
			},
		},
		{
			name:          "send-player-message",
			method:        http.MethodPost,
			path:          "/SendPlayerMessage",
			requestBody:   `{"SendType":"PlayerChat","Message":"Hello","UserID":"steam_76561198012345678"}`,
			responseBody:  postResponse(t, postResponses, "/SendPlayerMessage"),
			errorStatuses: errorStatuses,
			invoke: func(client *Client) error {
				_, err := client.SendPlayerMessage(context.Background(), &SendPlayerMessageRequest{
					SendType: "PlayerChat",
					Message:  "Hello",
					UserID:   contractPlayer,
				})
				return err
			},
		},
	}
}

func newContractServer(t *testing.T, endpoint contractEndpoint, status int) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertContractRequestMetadata(t, endpoint, r)
		assertContractRequestBody(t, endpoint, r)
		writeContractResponse(t, endpoint, status, w)
	}))
}

func assertContractRequestMetadata(t *testing.T, endpoint contractEndpoint, r *http.Request) {
	t.Helper()

	if r.Method != endpoint.method {
		t.Errorf("%s: method = %s, want %s", endpoint.name, r.Method, endpoint.method)
	}
	if want := contractPrefix + endpoint.path; r.URL.Path != want {
		t.Errorf("%s: path = %s, want %s", endpoint.name, r.URL.Path, want)
	}
	if got := r.Header.Get("Authorization"); got != "Bearer "+contractToken {
		t.Errorf("%s: Authorization = %q, want bearer token", endpoint.name, got)
	}
}

func assertContractRequestBody(t *testing.T, endpoint contractEndpoint, r *http.Request) {
	t.Helper()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		t.Errorf("%s: reading request body: %v", endpoint.name, err)
		return
	}
	if endpoint.requestBody == "" {
		if len(body) != 0 {
			t.Errorf("%s: unexpected request body: %s", endpoint.name, body)
		}
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Errorf("%s: unexpected Content-Type for empty body: %q", endpoint.name, got)
		}
		return
	}
	if got := r.Header.Get("Content-Type"); got != "application/json" {
		t.Errorf("%s: Content-Type = %q, want application/json", endpoint.name, got)
	}
	assertContractJSON(t, endpoint.name+" request", body, []byte(endpoint.requestBody))
}

func writeContractResponse(t *testing.T, endpoint contractEndpoint, status int, w http.ResponseWriter) {
	t.Helper()

	w.Header().Set("Content-Type", "application/json")
	if status != http.StatusOK {
		w.WriteHeader(status)
		_, _ = io.WriteString(w, `{"Error":{"Code":"ERROR_CODE","Message":"request failed","Details":{}}}`)
		return
	}
	if endpoint.responseBody == "" {
		t.Errorf("%s: missing response fixture", endpoint.name)
	}
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, endpoint.responseBody)
}

func assertContractJSON(t *testing.T, name string, got, want []byte) {
	t.Helper()

	var gotValue, wantValue any
	if err := json.Unmarshal(got, &gotValue); err != nil {
		t.Fatalf("%s: invalid JSON: %v", name, err)
	}
	if err := json.Unmarshal(want, &wantValue); err != nil {
		t.Fatalf("%s: invalid expected JSON: %v", name, err)
	}
	if !reflect.DeepEqual(gotValue, wantValue) {
		t.Fatalf("%s: got %s, want %s", name, got, want)
	}
}

func readContractFixture(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join("testdata", "contract", contractRevision, name)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read contract fixture %s: %v", path, err)
	}
	return strings.TrimSpace(string(data))
}

func readContractResponses(t *testing.T) map[string]string {
	t.Helper()

	var responses map[string]string
	data := readContractFixture(t, "post-success.json")
	if err := json.Unmarshal([]byte(data), &responses); err != nil {
		t.Fatalf("decode POST response fixtures: %v", err)
	}
	return responses
}

func postResponse(t *testing.T, responses map[string]string, path string) string {
	t.Helper()

	response, ok := responses[path]
	if !ok {
		t.Fatalf("missing POST response fixture for %s", path)
	}
	return response
}
