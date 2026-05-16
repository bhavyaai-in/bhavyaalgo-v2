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
    return proc


def kill_stale_port(port):
    try:
        result = subprocess.run(
            ["lsof", "-ti", f":{port}"],
            capture_output=True, text=True, timeout=5,
        )
        for line in result.stdout.strip().splitlines():
            pid = line.strip()
            if pid:
                os.kill(int(pid), signal.SIGKILL)
                print(f"Killed stale process {pid} on port {port}")
    except Exception as e:
        if "no such process" not in str(e).lower():
            pass


def wait_for_backend(port, timeout=15):
    import socket
    start = time.time()
    while time.time() - start < timeout:
        try:
            s = socket.create_connection(("localhost", port), timeout=2)
            s.close()
            return True
        except Exception:
            time.sleep(0.5)
    return False


def main():
    kill_stale_port(8080)
    time.sleep(1)

    if not os.path.isfile(AIR_PATH):
        print(f"Installing air to {AIR_PATH}...")
        run(f"{GO_BIN} install github.com/air-verse/air@latest", cwd=BACKEND_DIR, wait=True)

    print("Installing frontend deps...")
    run("npm install", cwd=FRONTEND_DIR, wait=True)

    print("Starting Go backend with air (port 8080)...")
    backend = run(f"{AIR_PATH}", cwd=BACKEND_DIR)

    if not wait_for_backend(8080):
        print("ERROR: Go backend failed to start. Check air output above.")
        backend.kill()
        sys.exit(1)

    print("Starting Vue dev server (port 5173)...")
    frontend = run("npm run dev", cwd=FRONTEND_DIR)

    print("Ready — open http://localhost:5173")

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
