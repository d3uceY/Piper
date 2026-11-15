"use strict";

// Maps Node's platform/arch pair onto the names GoReleaser puts in the
// release archives. Keep in sync with name_template in .goreleaser.yaml:
//
//   {{ .ProjectName }}_{{ title .Os }}_{{ if amd64 }}x86_64{{ else }}...
//
// so "linux" + "x64" becomes piper_Linux_x86_64.tar.gz.

const ARCH = {
  x64: "x86_64",
  arm64: "arm64",
  arm: "armv6",
};

const OS = {
  darwin: "Darwin",
  linux: "Linux",
  win32: "Windows",
};

// resolve returns the release asset name and how it is packed, or throws with
// a message worth showing to the user.
function resolve(platform = process.platform, arch = process.arch) {
  const os = OS[platform];
  const cpu = ARCH[arch];

  if (!os) {
    throw new Error(
      `piper does not ship a binary for platform "${platform}". ` +
        `Supported: ${Object.keys(OS).join(", ")}.`
    );
  }
  if (!cpu) {
    throw new Error(
      `piper does not ship a binary for architecture "${arch}". ` +
        `Supported: ${Object.keys(ARCH).join(", ")}.`
    );
  }

  const ext = platform === "win32" ? "zip" : "tar.gz";
  return {
    asset: `piper_${os}_${cpu}.${ext}`,
    archive: ext,
    binary: platform === "win32" ? "piper.exe" : "piper",
  };
}

module.exports = { resolve };
