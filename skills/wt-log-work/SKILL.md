---
name: wt-log-work
description: Use when the user wants to log, record or fill in work hours in Work Tracker (wt), e.g. "幫我記今天的工作", "log 2 hours of meetings", "補登昨天的工時", or describes what they worked on and asks to save it.
---

# Log work to Work Tracker

Adds work records to the user's Work Tracker through `scripts/wt.py` (Python 3, stdlib only). Every command prints JSON.

```bash
WT="python3 ~/.claude/skills/wt-log-work/scripts/wt.py"
```

## Steps

1. **Check the connection**: `$WT check`.
   - Exit code 2 (`"configured": false`) or a token error (`hint` mentions setup): tell the user to run
     `! python3 ~/.claude/skills/wt-log-work/scripts/wt.py setup` themselves. It asks for the site URL and an API
     token (created on the site's Profile page → API tokens) with hidden input. Do not ask them to paste the token
     into the chat. Run `check` again after they finish.
   - Connection error: show it and ask whether the URL is right.
2. **Load the choices**: `$WT options` gives the active category and project names.
3. **Turn the user's words into records**, one per piece of work:

   | Field | Rule |
   | - | - |
   | date | `YYYY-MM-DD`; default today. Resolve "昨天"/"週一" against today's date. |
   | category | Required. One of the listed categories. |
   | hours | Required. Over 0, at most 24, a multiple of 0.5. |
   | description | Optional; leave it empty when the user gives none. Otherwise keep the user's wording and language. |
   | project | Optional. One of the listed projects, or none. |

   Ask the user when a category or the hours can't be inferred with confidence. Never invent a category or project
   name; never round hours without saying so.
4. **Confirm before saving**: show a table (date, category, project, hours, description) and wait for the user's OK.
5. **Save** each record:
   ```bash
   $WT log --date 2026-10-02 --category "開發" --project "Work Tracker" --hours 1.5 --description "API token 功能"
   ```
   Use `--dry-run` to validate without saving. An unknown name returns `candidates`; pick from them or ask.
6. **Report** what was saved and the total hours. If one record failed, say which and why; the others are already saved.

## Other commands

- `$WT records --from 2026-09-28 --to 2026-10-04`: the user's records, e.g. to avoid logging the same work twice.
- `$WT --help` lists everything.
