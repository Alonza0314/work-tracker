---
name: integrate
description: Use when the user runs /integrate or asks to install, update or integrate the Work Tracker (wt) skills from this repo's skills/ folder into their Claude Code environment.
disable-model-invocation: true
---

# Install the Work Tracker skills

Installs every `wt-*` folder of the repo's `skills/` folder (the one holding this `.claude/`) into
`~/.claude/skills/`, replacing older copies.

1. **Install**:

   ```bash
   cd "$(git rev-parse --show-toplevel)/skills"
   mkdir -p ~/.claude/skills
   for src in wt-*/; do
       name=$(basename "$src")
       rm -rf ~/.claude/skills/"$name"
       cp -RL "$src" ~/.claude/skills/"$name"     # -L: symlinked scripts become real files
       find ~/.claude/skills/"$name" -name __pycache__ -prune -exec rm -rf {} +
       echo "installed $name"
   done
   ```

2. **Verify**: `ls -l ~/.claude/skills/wt-*/scripts/` shows a regular `wt.py` (not a symlink) in every skill, and
   `python3 ~/.claude/skills/wt-log-work/scripts/wt.py check` runs.
3. **Report** the installed skills and tell the user to reload Claude Code (restart it, or start a new session).
4. **Config**: if `check` exits with code 2 (not configured) or reports a token error, tell the user to run
   `! python3 ~/.claude/skills/wt-log-work/scripts/wt.py setup` themselves. It asks for the site URL and an API
   token (Profile page → API tokens) with hidden input. Do not ask them to paste the token into the chat.
