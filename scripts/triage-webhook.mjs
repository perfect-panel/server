import { createHmac } from "node:crypto";
import { readFileSync } from "node:fs";

const webhookUrl = process.env.BRIDGE_WEBHOOK_URL;
const webhookSecret = process.env.BRIDGE_WEBHOOK_SECRET;

if (!webhookUrl || !webhookSecret) {
  console.error("BRIDGE_WEBHOOK_URL and BRIDGE_WEBHOOK_SECRET are required");
  process.exit(1);
}

const context = JSON.parse(readFileSync(".automation/context.json", "utf8"));

const payload = {
  routeId: "perfect-panel-server-triage",
  eventType: "triage.issue",
  repo: "perfect-panel/server",
  source: "github-actions",
  trigger: {
    kind: process.env.GITHUB_EVENT_NAME || "unknown",
    eventName: process.env.GITHUB_EVENT_NAME || "unknown",
    eventAction: process.env.GITHUB_EVENT_ACTION || "",
  },
  context,
};

const rawBody = JSON.stringify(payload);
const signature = `sha256=${createHmac("sha256", webhookSecret).update(rawBody).digest("hex")}`;

const response = await fetch(webhookUrl, {
  method: "POST",
  headers: {
    "content-type": "application/json",
    "x-openclaw-signature-256": signature,
  },
  body: rawBody,
});

if (!response.ok) {
  const text = await response.text();
  console.error(`Webhook request failed: ${response.status} ${text}`);
  process.exit(1);
}

console.log(await response.text());
