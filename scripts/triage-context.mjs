import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const token = process.env.TRIAGE_TOKEN;
const eventName = process.env.GITHUB_EVENT_NAME || "unknown";
const eventPath = process.env.GITHUB_EVENT_PATH;
const repo = process.env.GITHUB_REPOSITORY || "perfect-panel/server";

if (!token) {
  console.error("TRIAGE_TOKEN is required");
  process.exit(1);
}

const event = eventPath && existsSync(eventPath) ? JSON.parse(readFileSync(eventPath, "utf8")) : {};

async function github(path) {
  const response = await fetch(`https://api.github.com${path}`, {
    headers: {
      Authorization: `Bearer ${token}`,
      Accept: "application/vnd.github+json",
      "User-Agent": "perfect-panel-triage",
    },
  });

  if (!response.ok) {
    throw new Error(`GitHub API ${path} failed: ${response.status} ${response.statusText}`);
  }

  return response.json();
}

const issues = await github(`/repos/${repo}/issues?state=open&per_page=50`);

const report = {
  generatedAt: new Date().toISOString(),
  eventName,
  trigger: {
    action: event.action ?? null,
    issue: event.issue
      ? {
          number: event.issue.number,
          title: event.issue.title,
          body: event.issue.body,
          url: event.issue.html_url,
          labels: (event.issue.labels || []).map((l) => l.name),
          user: event.issue.user?.login,
        }
      : null,
    comment: event.comment
      ? {
          id: event.comment.id,
          body: event.comment.body,
          url: event.comment.html_url,
          user: event.comment.user?.login,
          createdAt: event.comment.created_at,
        }
      : null,
  },
  openIssues: issues
    .filter((issue) => !issue.pull_request)
    .map((issue) => ({
      number: issue.number,
      title: issue.title,
      url: issue.html_url,
      labels: (issue.labels || []).map((l) => l.name),
      createdAt: issue.created_at,
    })),
};

mkdirSync(".automation", { recursive: true });
writeFileSync(join(".automation", "context.json"), `${JSON.stringify(report, null, 2)}\n`);
console.log(JSON.stringify(report, null, 2));
