package notify

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
	"google.golang.org/api/transport/cert"
	ghttp "google.golang.org/api/transport/http"
)

// Client プッシュ通知クライアント
type Client interface {
	Send(ctx context.Context, message *messaging.Message) (string, error)
}

// NewClient クライアントを作成する
func NewClient(ctx context.Context, enableLog bool) (Client, error) {
	opts, err := createClientOptions(ctx, enableLog)
	if err != nil {
		return nil, err
	}
	app, err := firebase.NewApp(ctx, nil, opts...)
	if err != nil {
		return nil, err
	}

	cli, err := app.Messaging(ctx)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func createClientOptions(ctx context.Context, enableLog bool) ([]option.ClientOption, error) {
	if !enableLog {
		return nil, nil
	}

	certSource, err := cert.DefaultSource()
	if err != nil {
		return nil, err
	}
	var baseTrans http.RoundTripper
	if certSource != nil {
		baseTrans = &http.Transport{
			TLSClientConfig: &tls.Config{
				GetClientCertificate: certSource,
			},
		}
	} else {
		baseTrans = http.DefaultTransport
	}
	var firebaseScopes = []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/datastore",
		"https://www.googleapis.com/auth/devstorage.full_control",
		"https://www.googleapis.com/auth/firebase",
		"https://www.googleapis.com/auth/identitytoolkit",
		"https://www.googleapis.com/auth/userinfo.email",
	}
	o := []option.ClientOption{option.WithScopes(firebaseScopes...)}
	trans, err := ghttp.NewTransport(ctx, baseTrans, o...)
	if err != nil {
		return nil, err
	}
	hc := &http.Client{
		Transport: &rt{
			base: trans,
		},
	}
	opt := option.WithHTTPClient(hc)
	return []option.ClientOption{opt}, nil
}

type rt struct {
	base http.RoundTripper
}

func (t *rt) RoundTrip(r *http.Request) (*http.Response, error) {
	dump, _ := httputil.DumpRequestOut(r, true)
	log.Printf("req:%s", dump)

	resp, err := t.base.RoundTrip(r)
	dump, _ = httputil.DumpResponse(resp, true)
	log.Printf("resp:%s", dump)

	return resp, err
}
