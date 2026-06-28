import { existsSync, mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const token = process.env.TRIAGE_TOKEN;
const eventName = process.env.GITHUB_EVENT_NAME ?? "unknown";
const eventPath = process.env.GITHUB_EVENT_PATH;
const repo = process.env.GITHUB_REPOSITORY ?? "perfect-panel/server";

if (!token) {
  console.error("TRIAGE_TOKEN is required");
  process.exit(1);
}

const event: Record<string, unknown> =
  eventPath && existsSync(eventPath)
    ? (JSON.parse(readFileSync(eventPath, "utf8")) as Record<string, unknown>)
    : {};

async function github<T>(path: string): Promise<T> {
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

  return response.json() as Promise<T>;
}

interface GitHubIssue {
  number: number;
  title: string;
  body: string;
  html_url: string;
  labels: { name: string }[];
  user: { login: string };
  created_at: string;
  pull_request?: unknown;
}

interface GitHubComment {
  id: number;
  body: string;
  html_url: string;
  user: { login: string };
  created_at: string;
}

const issues = await github<GitHubIssue[]>(`/repos/${repo}/issues?state=open&per_page=50`);

const triggerIssue = event.issue as GitHubIssue | undefined;
const triggerComment = event.comment as GitHubComment | undefined;

const report = {
  generatedAt: new Date().toISOString(),
  eventName,
  trigger: {
    action: (event.action as string) ?? null,
    issue: triggerIssue
      ? {
          number: triggerIssue.number,
          title: triggerIssue.title,
          body: triggerIssue.body,
          url: triggerIssue.html_url,
          labels: (triggerIssue.labels ?? []).map((l) => l.name),
          user: triggerIssue.user?.login,
        }
      : null,
    comment: triggerComment
      ? {
          id: triggerComment.id,
          body: triggerComment.body,
          url: triggerComment.html_url,
          user: triggerComment.user?.login,
          createdAt: triggerComment.created_at,
        }
      : null,
  },
  openIssues: issues
    .filter((issue) => !issue.pull_request)
    .map((issue) => ({
      number: issue.number,
      title: issue.title,
      url: issue.html_url,
      labels: (issue.labels ?? []).map((l) => l.name),
      createdAt: issue.created_at,
    })),
};

mkdirSync(".automation", { recursive: true });
writeFileSync(join(".automation", "context.json"), `${JSON.stringify(report, null, 2)}\n`);
console.log(JSON.stringify(report, null, 2));
