#!/usr/bin/env node
"use strict";

// The npm entry point. npm cannot publish a native executable, so the binary
// is downloaded beside this shim by install.js and spawned here. If the
// download did not happen (offline install, skipped postinstall), retry once
// before giving up.

const fs = require("node:fs");
const path = require("node:path");
const { spawn } = require("node:child_process");
const { resolve } = require("../lib/platform");

function binPath() {
  const target = resolve();
  return path.join(__dirname, "..", "vendor", `${process.platform}-${process.arch}`, target.binary);
}

function ensureBinary() {
  const binary = binPath();
  if (fs.existsSync(binary)) return binary;

  process.stderr.write("piper: binary missing, downloading...\n");
  try {
    require.main.require("../install.js");
  } catch {
    // install.js runs on require; the check below reports the real state.
  }
  return fs.existsSync(binary) ? binary : null;
}

const binary = ensureBinary();
if (!binary) {
  process.stderr.write(
    "piper: could not install the binary.\n" +
      "piper: check https://github.com/d3uceY/Piper/releases for a manual download.\n"
  );
  process.exit(1);
}

const child = spawn(binary, process.argv.slice(2), { stdio: "inherit" });

child.on("error", (err) => {
  process.stderr.write(`piper: ${err.message}\n`);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) process.kill(process.pid, signal);
  else process.exit(code ?? 1);
});
