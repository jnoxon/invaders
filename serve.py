#!/usr/bin/env python3
import http.server
import sys


class NoCacheHandler(http.server.SimpleHTTPRequestHandler):
    def end_headers(self):
        self.send_header("Cache-Control", "no-store")
        super().end_headers()


def main():
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 9090
    http.server.ThreadingHTTPServer(("0.0.0.0", port), NoCacheHandler).serve_forever()


if __name__ == "__main__":
    main()
