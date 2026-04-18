import type { ExtensionAPI, ExtensionContext } from "@mariozechner/pi-coding-agent";
import { withFileMutationQueue } from "@mariozechner/pi-coding-agent";
import { Type } from "@sinclair/typebox";
import * as fs from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import * as path from "node:path";

function findProjectPiDir(start: string): string {
  let current = path.resolve(start);
  while (true) {
    const candidate = path.join(current, ".pi");
    if (fs.existsSync(candidate) && fs.statSync(candidate).isDirectory()) {
      return candidate;
    }
    const parent = path.dirname(current);
    if (parent === current) {
      return path.join(path.resolve(start), ".pi");
    }
    current = parent;
  }
}

function defaultWorklogPath(ctx: ExtensionContext): string {
  return path.join(findProjectPiDir(ctx.cwd), "worklog", "devarch-v2.md");
}

function renderEntry(timestamp: string, params: {
  phase?: string;
  packet?: string;
  status?: string;
  summary: string;
  files?: string[];
}): string {
  const lines = ["", `## ${timestamp}`];
  if (params.phase) lines.push(`- phase: ${params.phase}`);
  if (params.packet) lines.push(`- packet: ${params.packet}`);
  if (params.status) lines.push(`- status: ${params.status}`);
  lines.push(`- summary: ${params.summary}`);
  if (params.files && params.files.length > 0) {
    lines.push(`- files: ${params.files.join(", ")}`);
  }
  return `${lines.join("\n")}\n`;
}

export default function worklogExtension(pi: ExtensionAPI): void {
  pi.registerCommand("worklog", {
    description: "Show the active DevArch V2 worklog path",
    handler: async (_args, ctx) => {
      ctx.ui.notify(`Worklog: ${defaultWorklogPath(ctx)}`, "info");
    },
  });

  pi.registerTool({
    name: "worklog_append",
    label: "Worklog Append",
    description: "Append a timestamped DevArch V2 worklog entry to the project worklog file.",
    promptSnippet: "Append short packet execution notes to the DevArch V2 worklog.",
    promptGuidelines: [
      "Use this tool after finishing a packet or phase note that should be preserved in the repo-local worklog.",
      "Keep entries short, factual, and tied to a phase, packet, or changed files whenever possible."
    ],
    parameters: Type.Object({
      summary: Type.String({ description: "Short factual note to append to the worklog." }),
      phase: Type.Optional(Type.String({ description: "Optional phase identifier such as phase-2 or P2." })),
      packet: Type.Optional(Type.String({ description: "Optional packet identifier such as P2-RSLV-003." })),
      status: Type.Optional(Type.String({ description: "Optional status label such as planned, in-progress, done, or blocked." })),
      files: Type.Optional(Type.Array(Type.String({ description: "Changed or relevant file path." }))),
      file: Type.Optional(Type.String({ description: "Optional custom worklog path. Defaults to .pi/worklog/devarch-v2.md at the project root." })),
    }),
    async execute(_toolCallId, params, _signal, _onUpdate, ctx) {
      const target = params.file ? path.resolve(ctx.cwd, params.file.replace(/^@/, "")) : defaultWorklogPath(ctx);
      const timestamp = new Date().toISOString();
      const entry = renderEntry(timestamp, params);

      return withFileMutationQueue(target, async () => {
        await mkdir(path.dirname(target), { recursive: true });
        let current = "# DevArch V2 Worklog\n\nChronological notes captured by the repo-local pi worklog extension.\n";
        try {
          current = await readFile(target, "utf8");
        } catch {
          // create the default heading on first write
        }
        const next = current.endsWith("\n") ? `${current}${entry}` : `${current}\n${entry}`;
        await writeFile(target, next, "utf8");
        return {
          content: [{ type: "text", text: `Appended worklog entry to ${path.relative(ctx.cwd, target) || target}` }],
          details: { path: target, timestamp },
        };
      });
    },
  });
}
