package pdrest_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	pdrest "github.com/TagamerStudio/pdrest-go"
)

func ExampleClient_GetVersion() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"Version":{"Major":0,"Minor":8,"Patch":1,"Version":"0.8.1"}}`)
	}))
	defer srv.Close()

	client, err := pdrest.NewClient(srv.URL, "bearer-token", pdrest.WithTimeout(30*time.Second))
	if err != nil {
		fmt.Println("client error:", err)
		return
	}
	defer func() { _ = client.Close() }()

	version, err := client.GetVersion(context.Background())
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	fmt.Printf("PalDefender %s", version.Version)
	// Output: PalDefender 0.8.1
}

func ExampleClient_GetPlayers() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"Meta":{"PlayerCount":1,"OnlineCount":1},"Players":[{"Name":"Alice","UserId":"steam_76561198012345678"}]}`)
	}))
	defer srv.Close()

	client, err := pdrest.NewClient(srv.URL, "bearer-token")
	if err != nil {
		fmt.Println("client error:", err)
		return
	}
	defer func() { _ = client.Close() }()

	players, err := client.GetPlayers(context.Background())
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	fmt.Printf("%d known players, first is %s", players.Meta.PlayerCount, players.Players[0].Name)
	// Output: 1 known players, first is Alice
}

func ExampleClient_GiveItems() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"Granted":{"Items":3}}`)
	}))
	defer srv.Close()

	client, err := pdrest.NewClient(srv.URL, "bearer-token")
	if err != nil {
		fmt.Println("client error:", err)
		return
	}
	defer func() { _ = client.Close() }()

	grant, err := client.GiveItems(context.Background(), "steam_76561198012345678", "Money", []any{"Money", 2})
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	fmt.Printf("granted %d item units", grant.Granted.Items)
	// Output: granted 3 item units
}

func ExampleClient_SummonPal() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"Summoned":{"Type":"Pal","PalID":"Foxparks","Level":12,"DamageMeter":true,"X":230,"Y":-486,"Z":4097}}`)
	}))
	defer srv.Close()

	client, err := pdrest.NewClient(srv.URL, "bearer-token")
	if err != nil {
		fmt.Println("client error:", err)
		return
	}
	defer func() { _ = client.Close() }()

	summoned, err := client.SummonPal(context.Background(), &pdrest.SummonPalRequest{
		PalID: "Foxparks",
		X:     230,
		Y:     -486,
		Z:     4097,
		Level: 12,
	})
	if err != nil {
		fmt.Println("request error:", err)
		return
	}
	fmt.Printf("summoned %s at %.0f,%.0f,%.0f", summoned.Summoned.PalID, summoned.Summoned.X, summoned.Summoned.Y, summoned.Summoned.Z)
	// Output: summoned Foxparks at 230,-486,4097
}

func ExampleClient_Alert() {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(w, `{"Error":{"Code":"INVALID_TOKEN","Message":"invalid token","Details":{}}}`)
	}))
	defer srv.Close()

	client, err := pdrest.NewClient(srv.URL, "wrong-token")
	if err != nil {
		fmt.Println("client error:", err)
		return
	}
	defer func() { _ = client.Close() }()

	_, err = client.Alert(context.Background(), "Restart now.")
	var apiErr *pdrest.APIError
	if errors.As(err, &apiErr) {
		fmt.Printf("API error: status %d on %s %s", apiErr.StatusCode, apiErr.Method, apiErr.Path)
	}
	// Output: API error: status 401 on POST /Alert
}
