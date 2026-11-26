package gateway

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/provider"
)

// fakeWSLNet makes nets the WSL networks Windows has an address on.
func fakeWSLNet(t *testing.T, cidrs ...string) {
	t.Helper()
	var nets []*net.IPNet
	for _, c := range cidrs {
		ip, n, err := net.ParseCIDR(c)
		if err != nil {
			t.Fatal(err)
		}
		n.IP = ip
		nets = append(nets, n)
	}
	old := WSLNets
	WSLNets = func() []*net.IPNet { return nets }
	t.Cleanup(func() { WSLNets = old })
}

// zaidu259434 on X: Codex in a WSL distro under NAT, pointed at Windows'
// address on the WSL network (172.17.32.1), failed to connect with
// sharing off. A request that reaches that address from the WSL network
// with the WSL key is served as one from this computer, its key the
// gateway's token (with the agent's id when the key has one); without the
// key, from elsewhere, or at another of this computer's addresses, it is
// judged as one from another machine and refused while magpie isn't
// shared.
func TestWSLCallerServedWithSharingOff(t *testing.T) {
	fresh(t)
	t.Setenv("MAGPIE_ADDR", "")
	fakeWSLNet(t, "172.17.32.1/20")
	key := WSLKey()
	if !strings.HasPrefix(key, wslKeyPrefix) || strings.HasPrefix(key, Token) {
		t.Fatalf("key %q", key)
	}
	if fi, err := os.Stat(wslKeyPath()); err != nil || fi.Mode().Perm()&0o077 != 0 && os.PathSeparator == '/' {
		t.Fatalf("key file: %v %v", fi, err)
	}
	var seen *http.Request
	h := lanGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r }))
	call := func(from, at string, set func(*http.Request)) (int, string) {
		seen = nil
		r := httptest.NewRequest("POST", "http://172.17.32.1:3425/v1/responses", nil)
		r.RemoteAddr = from
		ip, port, _ := net.SplitHostPort(at)
		p := 0
		for _, c := range port {
			p = p*10 + int(c-'0')
		}
		r = r.WithContext(context.WithValue(r.Context(), http.LocalAddrContextKey, &net.TCPAddr{IP: net.ParseIP(ip), Port: p}))
		if set != nil {
			set(r)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code, w.Body.String()
	}
	bearer := func(v string) func(*http.Request) {
		return func(r *http.Request) { r.Header.Set("Authorization", "Bearer "+v) }
	}
	const distro, win = "172.17.38.240:51000", "172.17.32.1:3425"

	if code, body := call(distro, win, bearer(key)); code != 200 || seen == nil {
		t.Fatalf("WSL key: %d %s", code, body)
	}
	if !local(seen) || seen.Header.Get("Authorization") != "Bearer "+Token {
		t.Errorf("served as %s with %q", seen.RemoteAddr, seen.Header.Get("Authorization"))
	}
	if code, _ := call(distro, win, bearer(key+"-codex")); code != 200 || agentOf(seen) != "codex" {
		t.Errorf("WSL key for codex: %d, agent %q", code, agentOf(seen))
	}
	if code, _ := call(distro, win, func(r *http.Request) { r.Header.Set("x-api-key", key) }); code != 200 || seen.Header.Get("x-api-key") != Token {
		t.Errorf("x-api-key: %d", code)
	}
	for name, c := range map[string]struct {
		from, at string
		set      func(*http.Request)
	}{
		"no key":            {distro, win, nil},
		"the token":         {distro, win, bearer(Token)},
		"a wrong key":       {distro, win, bearer(key[:len(key)-1] + "x")},
		"the key, extended": {distro, win, bearer(key + "x")},
		"another network":   {"192.168.1.20:5000", win, bearer(key)},
		"another address":   {distro, "192.168.1.5:3425", bearer(key)},
	} {
		if code, body := call(c.from, c.at, c.set); code != http.StatusForbidden || !strings.Contains(body, "isn't shared") {
			t.Errorf("%s: %d %s, want refused as another machine's", name, code, body)
		}
	}
}

// shuitunlulu_ on Discord: Claude Science in a WSL distro wasn't found or
// connected by magpie on Windows. Its daemon takes only an address
// (ANTHROPIC_BASE_URL, http only to loopback) and sends its Claude sign-in
// as its bearer, so under NAT it is given the relay in the distro with the
// WSL key as the address's first segment. From the WSL network that
// request is served as this computer's at the path without the key, its
// sign-in left as it was; the key in the path from elsewhere, or a wrong
// one, is no WSL request.
func TestWSLKeyInPath(t *testing.T) {
	fresh(t)
	t.Setenv("MAGPIE_ADDR", "")
	fakeWSLNet(t, "172.17.32.1/20")
	key := WSLKey()
	var seen *http.Request
	h := lanGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { seen = r }))
	call := func(from, path string) (int, string) {
		seen = nil
		r := httptest.NewRequest("POST", "http://172.17.32.1:3425"+path, nil)
		r.RemoteAddr = from
		r = r.WithContext(context.WithValue(r.Context(), http.LocalAddrContextKey, &net.TCPAddr{IP: net.ParseIP("172.17.32.1"), Port: 3425}))
		r.Header.Set("Authorization", "Bearer sk-ant-oat01-signin")
		r.Header.Set("X-Anthropic-Surface", "operon")
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code, w.Body.String()
	}
	const distro = "172.17.38.240:51000"
	if code, body := call(distro, "/"+key+"/v1/messages?beta=true"); code != 200 || seen == nil {
		t.Fatalf("key in the path: %d %s", code, body)
	}
	if !local(seen) || seen.URL.Path != "/v1/messages" || seen.URL.RawQuery != "beta=true" || seen.RequestURI != "/v1/messages?beta=true" {
		t.Errorf("served as %s at %q (%q)", seen.RemoteAddr, seen.URL.Path, seen.RequestURI)
	}
	if got := seen.Header.Get("Authorization"); got != "Bearer sk-ant-oat01-signin" || !claudeSignIn(seen) || agentOf(seen) != ScienceID {
		t.Errorf("sign-in %q, agent %q", got, agentOf(seen))
	}
	for name, c := range map[string]struct{ from, path string }{
		"no key":          {distro, "/v1/messages"},
		"a wrong key":     {distro, "/" + key[:len(key)-1] + "x/v1/messages"},
		"another network": {"192.168.1.20:5000", "/" + key + "/v1/messages"},
	} {
		if code, body := call(c.from, c.path); code != http.StatusForbidden || !strings.Contains(body, "isn't shared") {
			t.Errorf("%s: %d %s, want refused as another machine's", name, code, body)
		}
	}
}

// The gateway on loopback also listens at each address Windows has on a
// WSL network, comes or goes with it, and stops when sharing puts it on
// every address anyway. [::1] stands for the WSL network's address here.
func TestServeWSLListensBesideLoopback(t *testing.T) {
	fresh(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	t.Setenv("MAGPIE_ADDR", ln.Addr().String())
	if l6, err := net.Listen("tcp6", "[::1]:0"); err != nil {
		t.Skip("no IPv6 loopback")
	} else {
		l6.Close()
	}
	old, oldW := wslServeEverywhere, wslWatch
	wslServeEverywhere, wslWatch = true, 20*time.Millisecond
	t.Cleanup(func() { wslServeEverywhere, wslWatch = old, oldW })
	var nets atomic.Pointer[[]*net.IPNet]
	setNets := func(n []*net.IPNet) { nets.Store(&n) }
	setNets(nil)
	oldNets := WSLNets
	WSLNets = func() []*net.IPNet { return *nets.Load() }
	t.Cleanup(func() { WSLNets = oldNets })

	s := &Server{}
	s.ln = ln
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") })}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.serveWSL(ctx, srv); close(done) }()
	t.Cleanup(func() { cancel(); <-done; srv.Close() })

	at := "http://" + net.JoinHostPort("::1", Port()) + "/"
	answers := func() bool {
		c := &http.Client{Timeout: time.Second}
		res, err := c.Get(at)
		if err != nil {
			return false
		}
		res.Body.Close()
		return true
	}
	waitFor := func(want bool, what string) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for answers() != want {
			if time.Now().After(deadline) {
				t.Fatalf("%s: %s answers = %v", what, at, !want)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
	if answers() {
		t.Fatal("answering before any WSL network")
	}
	setNets([]*net.IPNet{{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)}})
	waitFor(true, "WSL network up")
	setNets(nil)
	waitFor(false, "WSL network gone")
	setNets([]*net.IPNet{{IP: net.ParseIP("::1"), Mask: net.CIDRMask(128, 128)}})
	waitFor(true, "WSL network back")
	// shared: the gateway listens on every address itself
	all, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { all.Close() })
	s.lnMu.Lock()
	s.ln = all
	s.lnMu.Unlock()
	waitFor(false, "shared on every address")
}

// H1_B1_K1 on X: a firewall kept WSL off vEthernet (WSL), and the distro
// reached Windows at its LAN address instead. At such a WSL via the WSL
// key is served from Windows' own address (a NATed distro's connection
// comes from there) or a WSL network, and refused from the LAN; an
// address that is no via is judged as before.
func TestWSLViaCallerServed(t *testing.T) {
	fresh(t)
	t.Setenv("MAGPIE_ADDR", "")
	fakeWSLNet(t, "172.17.32.1/20")
	oldVia := WSLVia
	WSLVia = func() []net.IP { return []net.IP{net.ParseIP("10.0.0.132")} }
	t.Cleanup(func() { WSLVia = oldVia })
	key := WSLKey()
	served := false
	h := lanGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { served = local(r) }))
	call := func(from, at string) int {
		served = false
		r := httptest.NewRequest("POST", "http://"+at+"/v1/responses", nil)
		r.RemoteAddr = from
		ip, _, _ := net.SplitHostPort(at)
		r = r.WithContext(context.WithValue(r.Context(), http.LocalAddrContextKey, &net.TCPAddr{IP: net.ParseIP(ip), Port: 3425}))
		r.Header.Set("Authorization", "Bearer "+key)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w.Code
	}
	for _, c := range []struct {
		from, at string
		ok       bool
	}{
		{"10.0.0.132:51000", "10.0.0.132:3425", true},     // NATed from the distro
		{"172.17.38.240:51000", "10.0.0.132:3425", true},  // from the WSL network
		{"10.0.0.20:51000", "10.0.0.132:3425", false},     // another machine on the LAN
		{"10.0.0.132:51000", "10.0.0.133:3425", false},    // an address that is no via
		{"172.17.38.240:51000", "172.17.32.1:3425", true}, // the WSL network, as before
	} {
		code := call(c.from, c.at)
		if (code == 200 && served) != c.ok {
			t.Errorf("%s to %s: %d served as local %v, want %v", c.from, c.at, code, served, c.ok)
		}
	}
}

// viaListener takes only connections from the address it listens at, or a
// WSL network; another machine's is closed as soon as it is taken.
func TestWSLViaListenerLetsInWindowsAndWSLOnly(t *testing.T) {
	fakeWSLNet(t, "172.17.32.1/20")
	tcp := func(s string) *net.TCPAddr { return &net.TCPAddr{IP: net.ParseIP(s), Port: 3425} }
	for _, c := range []struct {
		peer string
		ok   bool
	}{{"10.0.0.132", true}, {"172.17.38.240", true}, {"10.0.0.20", false}, {"192.168.1.5", false}} {
		if got := viaPeer(tcp(c.peer), tcp("10.0.0.132")); got != c.ok {
			t.Errorf("from %s: let in %v, want %v", c.peer, got, c.ok)
		}
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	v := viaListener{ln}
	defer v.Close()
	got := make(chan net.Conn, 1)
	go func() {
		if c, err := v.Accept(); err == nil {
			got <- c
		}
	}()
	c, err := net.Dial("tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	select {
	case s := <-got:
		s.Close() // from itself: let in
	case <-time.After(10 * time.Second):
		t.Fatal("a connection from the address listened at was not let in")
	}
}

// The gateway on loopback listens at a WSL via of this computer's as it
// is set, and lets it go when it is unset.
func TestServeWSLListensAtVia(t *testing.T) {
	fresh(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	t.Setenv("MAGPIE_ADDR", ln.Addr().String())
	if l6, err := net.Listen("tcp6", "[::1]:0"); err != nil {
		t.Skip("no IPv6 loopback")
	} else {
		l6.Close()
	}
	old, oldW := wslServeEverywhere, wslWatch
	wslServeEverywhere, wslWatch = true, time.Hour // only SetWSLVia's nudge looks again
	t.Cleanup(func() { wslServeEverywhere, wslWatch = old, oldW })
	fakeWSLNet(t)
	oldOwn := ownIP
	ownIP = func(ip net.IP) bool { return ip.Equal(net.IPv6loopback) }
	t.Cleanup(func() { ownIP = oldOwn; SetWSLVia(nil, nil) })

	s := &Server{}
	s.ln = ln
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") })}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { s.serveWSL(ctx, srv); close(done) }()
	t.Cleanup(func() { cancel(); <-done; srv.Close() })

	at := "http://" + net.JoinHostPort("::1", Port()) + "/"
	answers := func() bool {
		res, err := (&http.Client{Timeout: time.Second}).Get(at)
		if err != nil {
			return false
		}
		res.Body.Close()
		return true
	}
	waitFor := func(want bool, what string) {
		t.Helper()
		for deadline := time.Now().Add(10 * time.Second); answers() != want; time.Sleep(10 * time.Millisecond) {
			if time.Now().After(deadline) {
				t.Fatalf("%s: %s answers = %v", what, at, !want)
			}
		}
	}
	if answers() {
		t.Fatal("answering before any via")
	}
	if err := SetWSLVia(nil, []string{"::1", "10.9.9.9"}); err != nil { // the second isn't this computer's
		t.Fatal(err)
	}
	waitFor(true, "via set")
	if err := SetWSLVia([]string{"::1"}, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(true, "the one tried kept")
	if err := SetWSLVia(nil, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(false, "via unset")
}

// syncLogBuf captures log output safely across background goroutines.
type syncLogBuf struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *syncLogBuf) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *syncLogBuf) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

// freeLoopbackPort finds a TCP port free on both 127.0.0.1 and [::1].
func freeLoopbackPort(t *testing.T) string {
	t.Helper()
	l4, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	_, port, _ := net.SplitHostPort(l4.Addr().String())
	l6, err := net.Listen("tcp6", net.JoinHostPort("::1", port))
	l4.Close()
	if err != nil {
		t.Skip("no IPv6 loopback on paired port")
	}
	l6.Close()
	return port
}

// Stopping the gateway while serveWSL holds a WSL listener must drain cleanly
// without logging "requests still in flight; they are cut" on a double-closed
// listener (#7729), while a genuine request still in flight past shortDrain
// still logs and gets cut.
func TestWSLShutdownDoesNotLogCut(t *testing.T) {
	if l6, err := net.Listen("tcp6", "[::1]:0"); err != nil {
		t.Skip("no IPv6 loopback")
	} else {
		l6.Close()
	}

	startWithWSL := func(t *testing.T) (s *Server, port string, logs *syncLogBuf, stop context.CancelFunc, waitDone func() error) {
		t.Helper()
		fresh(t)
		port = freeLoopbackPort(t)
		t.Setenv("MAGPIE_ADDR", "127.0.0.1:"+port)

		oldServe, oldWatch := wslServeEverywhere, wslWatch
		wslServeEverywhere, wslWatch = true, time.Hour
		t.Cleanup(func() { wslServeEverywhere, wslWatch = oldServe, oldWatch })
		fakeWSLNet(t, "::1/128")

		logs = &syncLogBuf{}
		oldOut := log.Writer()
		log.SetOutput(logs)
		t.Cleanup(func() { log.SetOutput(oldOut) })

		s = New()
		stop, done := listening(t, s)
		var once sync.Once
		var doneErr error
		waitDone = func() error {
			once.Do(func() {
				select {
				case doneErr = <-done:
				case <-time.After(15 * time.Second):
					t.Fatal("ListenAndServe did not return")
				}
			})
			return doneErr
		}
		t.Cleanup(func() { stop(); waitDone() })

		at := "http://" + net.JoinHostPort("::1", port) + "/"
		client := &http.Client{
			Timeout:   time.Second,
			Transport: &http.Transport{DisableKeepAlives: true},
		}
		deadline := time.Now().Add(10 * time.Second)
		for {
			res, err := client.Get(at)
			if err == nil {
				res.Body.Close()
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("WSL listener %s never answered: %v", at, err)
			}
			time.Sleep(10 * time.Millisecond)
		}
		return s, port, logs, stop, waitDone
	}

	t.Run("serveWSL leaves active listener for srv.Shutdown", func(t *testing.T) {
		fresh(t)
		ln, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { ln.Close() })
		t.Setenv("MAGPIE_ADDR", ln.Addr().String())
		oldServe, oldWatch := wslServeEverywhere, wslWatch
		wslServeEverywhere, wslWatch = true, time.Hour
		t.Cleanup(func() { wslServeEverywhere, wslWatch = oldServe, oldWatch })
		fakeWSLNet(t, "::1/128")

		s := &Server{ln: ln}
		srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "ok") })}
		ctx, cancel := context.WithCancel(context.Background())
		done := make(chan struct{})
		go func() { s.serveWSL(ctx, srv); close(done) }()
		t.Cleanup(func() { cancel(); <-done; srv.Close() })

		at := "http://" + net.JoinHostPort("::1", Port()) + "/"
		client := &http.Client{Timeout: time.Second, Transport: &http.Transport{DisableKeepAlives: true}}
		for deadline := time.Now().Add(10 * time.Second); ; time.Sleep(10 * time.Millisecond) {
			if res, err := client.Get(at); err == nil {
				res.Body.Close()
				break
			}
			if time.Now().After(deadline) {
				t.Fatalf("WSL listener %s never answered", at)
			}
		}
		cancel()
		<-done
		shutCtx, shutCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutCancel()
		if err := srv.Shutdown(shutCtx); err != nil {
			t.Fatalf("srv.Shutdown returned %v, want nil", err)
		}
		if c, err := net.DialTimeout("tcp", net.JoinHostPort("::1", Port()), 200*time.Millisecond); err == nil {
			c.Close()
			t.Fatal("WSL listener stayed open after srv.Shutdown")
		}
	})

	t.Run("idle shutdown with WSL listener", func(t *testing.T) {
		_, port, logs, stop, waitDone := startWithWSL(t)
		stop()
		if err := waitDone(); err != nil {
			t.Fatalf("ListenAndServe returned %v", err)
		}
		if out := logs.String(); strings.Contains(out, "requests still in flight") {
			t.Fatalf("idle shutdown logged cut requests:\n%s", out)
		}
		if c, err := net.DialTimeout("tcp", net.JoinHostPort("::1", port), 200*time.Millisecond); err == nil {
			c.Close()
			t.Fatalf("WSL listener [::1]:%s stayed open after shutdown", port)
		}
	})

	t.Run("handed-off idle shutdown with WSL listener", func(t *testing.T) {
		s, port, logs, stop, waitDone := startWithWSL(t)
		s.HandOff()
		stop()
		if err := waitDone(); err != nil {
			t.Fatalf("ListenAndServe returned %v", err)
		}
		if out := logs.String(); strings.Contains(out, "requests still in flight") {
			t.Fatalf("handed-off idle shutdown logged cut requests:\n%s", out)
		}
		if c, err := net.DialTimeout("tcp", net.JoinHostPort("::1", port), 200*time.Millisecond); err == nil {
			c.Close()
			t.Fatalf("WSL listener [::1]:%s stayed open after handed-off shutdown", port)
		}
	})

	t.Run("in-flight request past drain still logs cut", func(t *testing.T) {
		oldShort := shortDrain
		shortDrain = 50 * time.Millisecond
		t.Cleanup(func() { shortDrain = oldShort })

		_, port, logs, stop, waitDone := startWithWSL(t)

		asked, cut, release := make(chan struct{}, 1), make(chan struct{}, 1), make(chan struct{})
		up := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			io.ReadAll(r.Body)
			select {
			case asked <- struct{}{}:
			default:
			}
			select {
			case <-r.Context().Done():
				select {
				case cut <- struct{}{}:
				default:
				}
			case <-release:
			}
		}))
		t.Cleanup(up.Close)
		t.Cleanup(func() { close(release) })
		if err := provider.Save(provider.Provider{ID: "slow", Name: "Slow", Key: "k", Chat: up.URL + "/v1", Models: []string{"m"}}); err != nil {
			t.Fatal(err)
		}

		reqDone := make(chan struct{})
		go func() {
			defer close(reqDone)
			req, _ := http.NewRequest("POST", "http://127.0.0.1:"+port+"/v1/chat/completions",
				strings.NewReader(`{"model":"slow/m","messages":[{"role":"user","content":"hi"}]}`))
			req.Header.Set("Content-Type", "application/json")
			if res, err := (&http.Client{Transport: &http.Transport{DisableKeepAlives: true}}).Do(req); err == nil {
				io.Copy(io.Discard, res.Body)
				res.Body.Close()
			}
		}()
		select {
		case <-asked:
		case <-time.After(10 * time.Second):
			t.Fatal("upstream was never reached")
		}

		stop()
		if err := waitDone(); err != nil {
			t.Fatalf("ListenAndServe returned %v", err)
		}
		select {
		case <-cut:
		case <-time.After(5 * time.Second):
			t.Fatal("in-flight upstream request was not canceled")
		}
		<-reqDone
		if out := logs.String(); !strings.Contains(out, "requests still in flight; they are cut") {
			t.Fatalf("expected cut-requests log when request outlives drain, got:\n%s", out)
		}
	})
}
