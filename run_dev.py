#!/usr/bin/env python3
"""Dev runner: builds frontend and runs backend with auto-reload."""

import os
import shutil
import signal
import subprocess
import sys
import time

ROOT = os.path.dirname(os.path.abspath(__file__))
FRONTEND_DIR = os.path.join(ROOT, "frontend")
BACKEND_DIR = os.path.join(ROOT, "backend")

AIR_PATH = os.path.expanduser("~/go/bin/air")
GO_BIN = shutil.which("go") or "/usr/local/go/bin/go"


def run(cmd, cwd, wait=False):
    proc = subprocess.Popen(
        cmd,
        cwd=cwd,
        shell=True,
        stdout=sys.stdout,
        stderr=sys.stderr,
    )
    if wait:
        proc.wait()
        return proc.returncode
    return proc


def kill_stale_port(port):
    import subprocess as sp
    try:
        result = sp.run(["lsof", "-ti", f":{port}"], capture_output=True, text=True, timeout=5)
        if result.stdout.strip():
            pids = result.stdout.strip().split()
            for pid in pids:
                os.kill(int(pid), signal.SIGKILL)
            print(f"Killed stale process(es) on port {port}")
    except Exception:
        pass


def wait_for_backend(port, timeout=10):
    start = time.time()
    while time.time() - start < timeout:
        try:
            import urllib.request
            urllib.request.urlopen(f"http://localhost:{port}/api/hello", timeout=2)
            return True
        except Exception:
            time.sleep(0.5)
    return False


def main():
    kill_stale_port(8080)

    if not os.path.isfile(AIR_PATH):
        print(f"Installing air to {AIR_PATH}...")
        run(f"{GO_BIN} install github.com/air-verse/air@latest", cwd=BACKEND_DIR, wait=True)

    print("Installing frontend deps...")
    run("npm install", cwd=FRONTEND_DIR, wait=True)

    print("Starting Vue dev server (port 5173)...")
    frontend = run("npm run dev", cwd=FRONTEND_DIR)

    print("Starting Go backend with air (port 8080)...")
    backend = run(f"{AIR_PATH}", cwd=BACKEND_DIR)

    if wait_for_backend(8080):
        print("Backend is ready — open http://localhost:5173")
    else:
        print("Warning: backend did not respond within timeout. Check for errors above.")

    def cleanup(sig=None, frame=None):
        print("\nShutting down...")
        for p in (backend, frontend):
            if p and p.poll() is None:
                p.terminate()
        for p in (backend, frontend):
            if p and p.poll() is None:
                try:
                    p.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    p.kill()
        sys.exit(0)

    signal.signal(signal.SIGINT, cleanup)
    signal.signal(signal.SIGTERM, cleanup)

    try:
        frontend.wait()
        backend.wait()
    except KeyboardInterrupt:
        cleanup()


if __name__ == "__main__":
    main()
