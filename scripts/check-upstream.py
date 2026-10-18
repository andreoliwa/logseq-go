"""Report changes since the reviewed upstream snapshot without changing the checkout."""

import fcntl
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import tempfile
import time

UPSTREAM = "https://github.com/aholstenson/logseq-go"
REF = "https://api.github.com/repos/aholstenson/logseq-go/git/ref/heads/main"
SHA = re.compile(r"[0-9a-f]{40}\Z")


def warn(message):
    print("warning: upstream check: " + message)


def read_state(path):
    try:
        state = json.loads(path.read_text())
        if not isinstance(state, dict):
            return {}
        return state
    except (OSError, ValueError):
        return {}


def write_state(path, state):
    # Use a unique sibling and replace atomically; readers never see partial JSON.
    name = None
    try:
        with tempfile.NamedTemporaryFile(mode="w", dir=path.parent, delete=False) as out:
            name = out.name
            json.dump(state, out)
        os.replace(name, path)
    finally:
        if name is not None:
            try:
                os.unlink(name)
            except FileNotFoundError:
                pass


def notice(state, reviewed):
    observed = state.get("observed", "")
    if state.get("reviewed") == reviewed and isinstance(observed, str) and SHA.fullmatch(observed) and observed != reviewed:
        print("Upstream differs from the reviewed snapshot: " + UPSTREAM + "/compare/" + reviewed + "..." + observed)


def lookup():
    # -q disables curlrc, avoiding hidden credentials or interactive host settings.
    process = subprocess.Popen(
        ["curl", "-q", "--fail", "--silent", "--show-error", "--connect-timeout", "2", "--max-time", "4", REF],
        stdout=subprocess.PIPE, stderr=subprocess.PIPE, start_new_session=True,
    )
    try:
        stdout, stderr = process.communicate(timeout=6)
    except subprocess.TimeoutExpired:
        os.killpg(process.pid, signal.SIGKILL)
        process.communicate()
        raise ValueError("lookup timed out") from None
    if process.returncode:
        # Do not echo external stderr, which can contain host-specific data.
        raise ValueError("lookup failed; retry after the cache interval")
    observed = json.loads(stdout)["object"]["sha"]
    if not isinstance(observed, str) or not SHA.fullmatch(observed):
        raise ValueError("unexpected ref response")
    return observed


def main():
    if os.environ.get("CI"):
        return 0
    try:
        reviewed = Path(__file__).with_name("upstream-reviewed.sha").read_text().strip()
        if not SHA.fullmatch(reviewed):
            raise ValueError("invalid reviewed snapshot")
        interval = int(os.environ.get("UPSTREAM_CHECK_INTERVAL_SECONDS", "86400"))
        if interval < 0:
            raise ValueError("interval must be nonnegative")
        home = os.environ.get("HOME")
        cache = os.environ.get("XDG_CACHE_HOME") or (str(Path(home) / ".cache") if home else None)
        if not cache:
            raise ValueError("HOME or XDG_CACHE_HOME is required for throttling")
        directory = Path(cache) / "logseq-go"
        directory.mkdir(parents=True, exist_ok=True)
        path = directory / "upstream.json"
        with (directory / "upstream.lock").open("a") as lock:
            try:
                # Kernel-owned locks are released on exit, including crashes. A
                # stale lock file is harmless and cannot suppress future checks.
                fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
            except BlockingIOError:
                notice(read_state(path), reviewed)
                return 0
            state = read_state(path)
            if state.get("reviewed") != reviewed:
                state = {"reviewed": reviewed}
            now = time.time()
            attempt = state.get("attempt")
            if interval and isinstance(attempt, (int, float)) and 0 <= now - attempt < interval:
                notice(state, reviewed)
                return 0
            state["attempt"] = now
            # Persist the attempt before networking: failures stay throttled.
            write_state(path, state)
            try:
                state["observed"] = lookup()
                state["success"] = now
                write_state(path, state)
            except (OSError, ValueError, KeyError, TypeError) as error:
                warn(str(error))
            notice(state, reviewed)
    except (OSError, ValueError) as error:
        warn(str(error))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
