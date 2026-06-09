"""mitmproxy addon: log every host the CLI routes through the proxy.

Logs at CONNECT time (http_connect) so HTTPS flows are recorded even when the
client later rejects the server certificate — which is exactly the case we test.
Also logs plain-HTTP requests. Host list is written to $PROXY_CONNECT_LOG.
"""
import os

LOG = os.environ.get("PROXY_CONNECT_LOG", "/tmp/proxy-connects.txt")


class ConnectLogger:
    def _write(self, host):
        with open(LOG, "a") as f:
            f.write(host + "\n")

    def http_connect(self, flow):
        self._write(flow.request.host)

    def request(self, flow):
        self._write(flow.request.pretty_host)


addons = [ConnectLogger()]
