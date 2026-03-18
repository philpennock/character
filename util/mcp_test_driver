#!/usr/bin/env python3
"""Interactive REPL for testing the character agent MCP stdio server.

Builds the binary if it is missing, performs the MCP handshake, then
presents a prompt where you can call any tool by name.  Every JSON-RPC
frame that goes in or comes out is traced to stderr so you can see exactly
what the protocol looks like.

Usage at the prompt
    list                         list available tools and descriptions
    <tool> [json | key=val ...]  call a tool
    help                         show this help
    quit / exit / Ctrl-D         exit
"""

import json
import shlex
import subprocess
import sys
from pathlib import Path

try:
    REPO_ROOT = [d for d in Path(__file__).absolute().parents if (d / '.git').exists()][0]
except IndexError as e:
    raise Exception(f'script {__file__} is not inside a git repository') from e
BINARY = REPO_ROOT / "character"

_id_seq = 0


# ── colour helpers ────────────────────────────────────────────────────────────

def _colour(code: str, s: str, stream) -> str:
    if stream.isatty():
        return f"\033[{code}m{s}\033[0m"
    return s


def _err(code: str, s: str) -> str:
    return _colour(code, s, sys.stderr)


def _out(code: str, s: str) -> str:
    return _colour(code, s, sys.stdout)


def cyan(s: str) -> str:   return _err("36", s)
def yellow(s: str) -> str: return _err("33", s)
def bold_err(s: str) -> str: return _err("1", s)
def bold(s: str) -> str:   return _out("1", s)
def red(s: str) -> str:    return _out("31", s)
def dim(s: str) -> str:    return _out("2", s)


def eprint(*args, **kwargs):
    print(*args, **kwargs, file=sys.stderr)


# ── MCP framing ───────────────────────────────────────────────────────────────

def _next_id() -> int:
    global _id_seq
    _id_seq += 1
    return _id_seq


def _frame(obj: dict) -> bytes:
    return json.dumps(obj, separators=(",", ":")).encode() + b"\n"


def send(proc, obj: dict):
    eprint(cyan(f">>> {json.dumps(obj)}"))
    proc.stdin.write(_frame(obj))
    proc.stdin.flush()


def recv(proc) -> dict | None:
    line = proc.stdout.readline()
    if not line:
        return None
    obj = json.loads(line)
    eprint(yellow(f"<<< {json.dumps(obj)}"))
    return obj


# ── build / launch ────────────────────────────────────────────────────────────

def ensure_binary():
    if not BINARY.exists():
        eprint(bold_err("character binary not found — building..."))
        result = subprocess.run(
            ["go", "build", "-o", str(BINARY), "."],
            cwd=REPO_ROOT,
        )
        if result.returncode != 0:
            sys.exit("go build failed")
        eprint(bold_err(f"Built {BINARY}"))


def launch() -> subprocess.Popen:
    # Child stderr inherits ours so any server-side error messages appear
    # naturally mixed in with our own stderr traces.
    return subprocess.Popen(
        [str(BINARY), "agent", "mcp"],
        stdin=subprocess.PIPE,
        stdout=subprocess.PIPE,
        cwd=REPO_ROOT,
    )


# ── MCP session ───────────────────────────────────────────────────────────────

def do_initialize(proc) -> dict:
    send(proc, {
        "jsonrpc": "2.0",
        "id": _next_id(),
        "method": "initialize",
        "params": {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "mcp_test_driver", "version": "0"},
        },
    })
    resp = recv(proc)
    info = resp.get("result", {}).get("serverInfo", {})
    eprint(bold_err(
        f"Connected: {info.get('name', '?')} {info.get('version', '')}"))
    # notification — no response expected
    send(proc, {"jsonrpc": "2.0",
                "method": "notifications/initialized", "params": {}})
    return resp


def get_tools(proc) -> list[dict]:
    send(proc, {"jsonrpc": "2.0", "id": _next_id(),
                "method": "tools/list", "params": {}})
    resp = recv(proc)
    return resp.get("result", {}).get("tools", [])


def call_tool(proc, name: str, arguments: dict) -> dict | None:
    send(proc, {
        "jsonrpc": "2.0",
        "id": _next_id(),
        "method": "tools/call",
        "params": {"name": name, "arguments": arguments},
    })
    return recv(proc)


# ── argument parsing ──────────────────────────────────────────────────────────

def parse_args(text: str) -> dict:
    """Accept a JSON object literal or space-separated key=value pairs."""
    text = text.strip()
    if not text:
        return {}
    if text.startswith("{"):
        return json.loads(text)

    result = {}
    for token in shlex.split(text):
        if "=" not in token:
            continue
        key, _, val = token.partition("=")
        # coerce obvious scalars so schemas stay happy
        if val.lower() == "true":
            result[key] = True
        elif val.lower() == "false":
            result[key] = False
        else:
            try:
                result[key] = int(val)
            except ValueError:
                result[key] = val
    return result


# ── REPL ──────────────────────────────────────────────────────────────────────

HELP = """
Commands
  list                         list tools and their descriptions
  <tool>                       call a tool with no arguments
  <tool> {"key": "val", ...}   call a tool with a raw JSON object
  <tool> key=val key=val ...   call a tool with keyword arguments
  help                         this message
  quit / exit / Ctrl-D         exit

Key=value notes
  Booleans: exact=true  exact=false
  Integers: decimal=10003  (parsed automatically)
  Strings:  all other values, quote with shell rules if they contain spaces

Examples
  unicode_lookup_char char=✓
  unicode_search query=snowman
  unicode_lookup_codepoint codepoint=U+2603
  unicode_lookup_name name="CHECK MARK" exact=true
  unicode_browse_block block=Dingbats
  unicode_list_blocks
  unicode_emoji_flag country_code=GB
  unicode_transform type=fraktur text="Hello world"
""".strip()


def _print_result(resp: dict):
    result = resp.get("result", {})
    is_error = result.get("isError", False)
    for item in result.get("content", []):
        text = item.get("text", "")
        if is_error:
            print(red(f"Error: {text}"))
        else:
            try:
                parsed = json.loads(text)
                print(json.dumps(parsed, indent=2, ensure_ascii=False))
            except json.JSONDecodeError:
                print(text)


def _build_completions(tools: list[dict]) -> tuple[set[str], dict[str, list[str]], dict[str, list[str]]]:
    """Extract tool names, per-tool argument keys, and per-key enum values
    from the tools/list response for tab-completion."""
    commands = {"list", "help", "quit", "exit"}
    tool_names: set[str] = set()
    tool_args: dict[str, list[str]] = {}    # tool → ["key=", ...]
    arg_enums: dict[str, list[str]] = {}    # "tool:key" → ["val1", ...]
    for t in tools:
        name = t["name"]
        tool_names.add(name)
        schema = t.get("inputSchema", {})
        props = schema.get("properties", {})
        keys = sorted(props.keys())
        tool_args[name] = [k + "=" for k in keys]
        for k, v in props.items():
            if "enum" in v:
                arg_enums[f"{name}:{k}"] = v["enum"]
    return commands | tool_names, tool_args, arg_enums


def _make_completer(commands: set[str], tool_names: set[str],
                    tool_args: dict[str, list[str]],
                    arg_enums: dict[str, list[str]]):
    """Return a readline completer function with closure over tool metadata."""

    def completer(text: str, state: int) -> str | None:
        import readline
        buf = readline.get_line_buffer()
        begin = readline.get_begidx()

        if begin == 0:
            # First word — complete command / tool names.
            matches = sorted(c for c in commands if c.startswith(text))
        else:
            # After the first word — figure out which tool is being invoked.
            first_word = buf[:begin].split()[0] if buf[:begin].strip() else ""
            if first_word in tool_names:
                # If text contains '=' we're completing a value.
                if "=" in text:
                    key, _, partial = text.partition("=")
                    enum_key = f"{first_word}:{key}"
                    if enum_key in arg_enums:
                        prefix = key + "="
                        matches = sorted(
                            prefix + v
                            for v in arg_enums[enum_key]
                            if v.startswith(partial)
                        )
                    else:
                        matches = []
                else:
                    # Complete argument keys (as "key=").
                    matches = sorted(
                        k for k in tool_args.get(first_word, [])
                        if k.startswith(text)
                    )
            else:
                matches = []

        return matches[state] if state < len(matches) else None

    return completer


def repl(proc, tools: list[dict]):
    tool_names = {t["name"] for t in tools}

    print()
    print(bold("character MCP test driver"))
    print(dim('Type "list" to see tools, "help" for usage, Ctrl-D to exit.'))
    print(dim("Tab-completion is available for commands, tools, and arguments."))
    print()

    try:
        import readline
    except ImportError:
        pass
    else:
        all_words, tool_args, arg_enums = _build_completions(tools)
        readline.set_completer(_make_completer(all_words, tool_names,
                                               tool_args, arg_enums))
        readline.set_completer_delims(" ")
        readline.parse_and_bind("tab: complete")

    while True:
        try:
            line = input("mcp> ").strip()
        except (EOFError, KeyboardInterrupt):
            print()
            break

        if not line:
            continue
        if line in ("quit", "exit"):
            break
        if line == "help":
            print(HELP)
            continue
        if line == "list":
            for t in tools:
                print(f"  {bold(t['name'])}: {dim(t['description'])}")
            continue

        parts = line.split(None, 1)
        name = parts[0]
        rest = parts[1] if len(parts) > 1 else ""

        if name not in tool_names:
            print(red(f"Unknown tool '{name}'."), dim("(type 'list' to see tools)"))
            continue

        try:
            arguments = parse_args(rest)
        except (json.JSONDecodeError, ValueError) as exc:
            print(red(f"Could not parse arguments: {exc}"))
            continue

        resp = call_tool(proc, name, arguments)
        if resp is None:
            print(red("Server closed connection unexpectedly."))
            break

        _print_result(resp)


# ── main ──────────────────────────────────────────────────────────────────────

def main():
    ensure_binary()
    proc = launch()
    try:
        do_initialize(proc)
        tools = get_tools(proc)
        repl(proc, tools)
    finally:
        proc.stdin.close()
        try:
            proc.wait(timeout=3)
        except subprocess.TimeoutExpired:
            proc.kill()


if __name__ == "__main__":
    main()
