package notify

import (
	"context"
	"crypto/tls"
	"errors"
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

type client struct {
	c     *messaging.Client
	reset func()
}

func (c *client) Send(ctx context.Context, message *messaging.Message) (string, error) {
	c.reset()
	return c.c.Send(ctx, message)
}

// NewClient クライアントを作成する
func NewClient(ctx context.Context, enableLog bool) (Client, error) {
	opts, reset, err := createClientOptions(ctx, enableLog)
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

	return &client{
		c:     cli,
		reset: reset,
	}, nil
}

func createClientOptions(ctx context.Context, enableLog bool) ([]option.ClientOption, func(), error) {
	if !enableLog {
		return nil, nil, nil
	}

	certSource, err := cert.DefaultSource()
	if err != nil {
		return nil, nil, err
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
		return nil, nil, err
	}

	rt := &rt{
		base: trans,
	}
	hc := &http.Client{
		Transport: rt,
	}
	opt := option.WithHTTPClient(hc)
	return []option.ClientOption{opt}, rt.reset, nil
}

type rt struct {
	base http.RoundTripper
	err  error
}

func (t *rt) RoundTrip(r *http.Request) (*http.Response, error) {
	if t.err != nil {
		return nil, t.err
	}

	dump, _ := httputil.DumpRequestOut(r, true)
	log.Printf("req:%s", dump)

	resp, err := t.base.RoundTrip(r)
	dump, _ = httputil.DumpResponse(resp, true)
	log.Printf("resp:%s", dump)

	if err != nil {
		return resp, err
	}

	// ServiceUnavailableならリトライさせない
	if resp.StatusCode == http.StatusServiceUnavailable {
		t.err = errors.New("ServiceUnavailable")
		return nil, t.err
	}

	return resp, nil
}

func (t *rt) reset() {
	t.err = nil
}
