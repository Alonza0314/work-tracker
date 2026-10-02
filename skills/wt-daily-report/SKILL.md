---
name: wt-daily-report
description: Use when the user prepares for a daily meeting / stand-up / 晨會 and wants to know what they did on the previous workday according to Work Tracker (wt), e.g. "整理昨天的工作給 daily", "what did I do yesterday".
---

# Daily meeting report from Work Tracker

Summarizes the user's work records of the previous workday through `scripts/wt.py` (Python 3, stdlib only).

```bash
WT="python3 ~/.claude/skills/wt-daily-report/scripts/wt.py"
```

## Steps

1. **Check the connection**: `$WT check`.
   - Exit code 2 (`"configured": false`) or a token error (`hint` mentions setup): tell the user to run
     `! python3 ~/.claude/skills/wt-daily-report/scripts/wt.py setup` themselves. It asks for the site URL and an
     API token (created on the site's Profile page → API tokens) with hidden input. Do not ask them to paste the
     token into the chat. Run `check` again after they finish.
2. **Fetch the day**: `$WT daily`. It picks the workday before today, skipping weekends and holidays and counting
   makeup workdays from the Work Tracker calendar. If the user names a day, use `$WT daily --date YYYY-MM-DD`.
3. **Write the report** in the user's language:

   ```
   **<date> (<weekday>)** — <totalHours> 小時

   **<project or category>**
   - <description> (<hours>h)
   ```

   - Group by project; records without a project are grouped by category.
   - Keep each description's meaning; shorten only long ones. Multi-line descriptions may become sub-bullets.
   - A record with an empty description is listed by its category name, e.g. `- 會議 (1h)`; one with neither is listed as `- (未填寫) (<hours>h)`.
   - Report only what is in the records. Do not add a plan for today or anything else the records don't say.
4. **No records** (`records` is empty): say that nothing was logged on that date and offer to log it with the
   `wt-log-work` skill.
