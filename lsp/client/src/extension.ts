import * as path from "path";
import * as vscode from "vscode";
import { watch, writeFileSync } from "fs";
import { exec } from "child_process";
import * as http from "http";
import {
  LanguageClient,
  LanguageClientOptions,
  ServerOptions,
  TransportKind,
} from "vscode-languageclient/node";
import * as fs from "fs";
import { tmpdir } from "os";
import { parse } from "./grammar/lib/parser/parsing";

interface Node {
  id: number;
  name: string;
  type: string;
  alias?: string;
  mime_type?: string;
  channel_name?: string;
}

interface Link {
  source: number;
  target: number;
  type?: string;
}

const default_struct = {
  nodes: [] as Node[],
  links: [] as Link[],
};

let outJson = { ...default_struct };
let port = 3001;  // dynamically assigned based on availability
const basePort = 3001;  // Starting port for search
let client: LanguageClient;
let panel: vscode.WebviewPanel;
let terminal: vscode.Terminal | undefined;
let tmpterminal: vscode.Terminal | undefined;
let httpServerProcess: any = null;
const supportedExtensions = new Set<string>([".dse"]);
const isCodespace = vscode.env.remoteName === "codespaces";
let astYamlPath: string = "";
let simulationYamlPath: string = "";
let cdDirPath: string = "";
const checkInterval = 1000;
const timeout = 30000;
const envVars: Record<string, string> = {};
let dseDirPath: string;
const tmpPreBuild = "pre_build_completed";
const tmpPreRun = "pre_run_completed";
const tmpPreClean = "pre_clean_completed";
const tmpPreCleanall = "pre_cleanall_completed";
const tmpSimRun = "sim_run_completed";
const tmpBuild = "build_completed";
const daemonWorkspace = "/var/lib/docker/codespacemount/workspace";

export function activate(context: vscode.ExtensionContext) {
  const diagnosticCollection = vscode.languages.createDiagnosticCollection("dse");
  diagnosticCollection.clear();
  context.subscriptions.push(diagnosticCollection);
  vscode.workspace.onDidChangeTextDocument((event) => {
    const document = event.document;
    if (document.languageId === "dse") {
      const text = document.getText();
      const diagnostics = parse(text);
      diagnosticCollection.clear();
      if (Array.isArray(diagnostics)) {
        diagnosticCollection.set(document.uri, diagnostics);
      }
    }
  });

  let activeEditor = vscode.window.activeTextEditor;
  const extPath = vscode.extensions.getExtension("dse.dse")!.extensionPath;

  if (isCodespace) {
    generateContainerHTML(
      path.join(extPath, "ast_dag", "ast.html"),
      process.env.CODESPACE_NAME,
    );
  }
  const switchPanel = async (isSideBySide: boolean) => {
    await vscode.window.withProgress(
      {
        location: vscode.ProgressLocation.Notification,
        title: "Loading Live Preview",
        cancellable: false,
      },
      async (progress) => {
        activeEditor = vscode.window.activeTextEditor;
        if (activeEditor && activeEditor.document.languageId === "dse") {
          const filePath = activeEditor.document.uri.fsPath;

          progress.report({ message: "Converting DSL to AST..." });
          const convStatus = await dslToAstConvertion(filePath, extPath);

          if (convStatus === true) {
            progress.report({ message: "Starting HTTP server..." });
            const status = await processAndServeFile(extPath);

            if (status === true) {
              progress.report({ message: "Opening preview panel..." });
              panel?.dispose();
              panel = vscode.window.createWebviewPanel(
                "livePreview",
                "DSE Live Preview",
                isSideBySide
                  ? vscode.ViewColumn.Beside // Open in the side-by-side panel
                  : vscode.ViewColumn.Active, // Open in a single panel
                {
                  enableScripts: true,
                  retainContextWhenHidden: false,
                },
              );

              // Kill the server when the panel is disposed
              panel.onDidDispose(() => {
                console.log("[INFO] Panel closed, killing HTTP server on port " + port);
                if (httpServerProcess) {
                  httpServerProcess.kill();
                  httpServerProcess = null;
                } else {
                  killProcess(port);
                }
              });

              let url = "";
              if (isCodespace) {
                url = `https://${process.env.CODESPACE_NAME}-${port}.app.github.dev/ast.html?t=${new Date().getTime()}`;
              } else {
                url = `http://127.0.0.1:${port}/ast.html?t=${new Date().getTime()}`;
              }
              panel.webview.html = getWebviewContent(url);
              let debounceTimer: NodeJS.Timeout;
              const debounceDelay = 1000;
              watch(filePath, async (eventType, filename) => {
                if (eventType === "change") {
                  const status = await dslToAstConvertion(filePath, extPath);
                  if (status === true) {
                    updateD3InputFile(extPath);
                    clearTimeout(debounceTimer);
                    debounceTimer = setTimeout(() => {
                      const cacheBustedUrl = `${url}?t=${new Date().getTime()}`;
                      panel.webview.html = getWebviewContent(cacheBustedUrl);
                      panel.webview.postMessage("refresh");
                    }, debounceDelay);
                  }
                }
              });
              vscode.window.showInformationMessage(
                `Live View created. Listening changes in file ${filePath}`,
              );
            }
          }
        }
      }
    );
  };

  context.subscriptions.push(
    vscode.commands.registerCommand("livePreview.toggle", () => {
      const editor = vscode.window.activeTextEditor;
      if (editor) {
        const filePath = editor.document.uri.fsPath;
        const activeFileExt = path.extname(filePath);
        if (supportedExtensions.has(activeFileExt)) {
          if (validateDiagnostics(editor, diagnosticCollection)) {
            switchPanel(false); // Open preview in the active panel
          } else {
            vscode.window.showErrorMessage(
              `This file contains error(s). Please fix them before proceeding.`,
            );
          }
        } else {
          vscode.window.showWarningMessage(
            `File extension ${activeFileExt} is NOT supported.`,
          );
        }
      }
    }),
  );

  context.subscriptions.push(
    vscode.commands.registerCommand("livePreview.toggleSideBySide", () => {
      const editor = vscode.window.activeTextEditor;
      if (editor) {
        const filePath = editor.document.uri.fsPath;
        const activeFileExt = path.extname(filePath);
        if (supportedExtensions.has(activeFileExt)) {
          if (validateDiagnostics(editor, diagnosticCollection)) {
            switchPanel(true); // Open preview in the side-by-side panel
          } else {
            vscode.window.showErrorMessage(
              `This file contains error(s). Please fix them before proceeding.`,
            );
          }
        } else {
          vscode.window.showWarningMessage(
            `File extension ${activeFileExt} is NOT supported.`,
          );
        }
      }
    }),
  );

  const build_cmd = vscode.commands.registerCommand("Build", () => {
    clearTempFiles();
    terminal = terminalSetup(terminal);
    const editor = vscode.window.activeTextEditor;
    if (editor) {
      const [filePath, activeFileExt, activeFileName, activeFileDirPath] =
        getActiveFileInfo(editor);
      cdDirPath = isCodespace
        ? activeFileDirPath
        : convertToMntPath(activeFileDirPath.replace(/\\/g, "/"));
      const genSimulationPath = path.join(activeFileDirPath, "out/simulation.yaml");
      const genTaskfilePath = path.join(activeFileDirPath, "out/Taskfile.yml");
      const astJsonPath = path.join(
        activeFileDirPath, 'out',
        activeFileName + ".json",
      );
      const astOutputPath = isCodespace
        ? astJsonPath
        : convertToMntPath(astJsonPath.replace(/\\/g, "/"));
      if (supportedExtensions.has(activeFileExt)) {
        if (validateDiagnostics(editor, diagnosticCollection)) {
          terminal?.sendText(`cd ${cdDirPath}`);
          tmpterminal = terminalSetup(tmpterminal);
          astYamlPath = path.join(activeFileDirPath, 'out', activeFileName + ".yaml");
          astYamlPath = isCodespace
            ? astYamlPath
            : convertToMntPath(astYamlPath.replace(/\\/g, "/"));
          removeFile(astYamlPath);
          removeFile(genTaskfilePath);
          removeFile(genSimulationPath);

          //if 'pre_build.sh' is present it gets executed first.
          const execFile = "pre_build.sh";
          const tmpPath: string = path.join(tmpdir(), tmpPreBuild);
          const preBuildCompletionStatusFile = isCodespace
            ? tmpPath
            : convertToMntPath(tmpPath.replace(/\\/g, "/"));
          const preBuildPath = path.join(activeFileDirPath, execFile);
          if (fs.existsSync(preBuildPath)) {
            terminal?.sendText(
              `sh ${execFile} && touch ${preBuildCompletionStatusFile}`,
            );
            waitForFile(tmpPath, () => {
              console.log(`executing ${execFile}`);
              build(
                filePath,
                astOutputPath,
                extPath,
                activeFileDirPath,
                genSimulationPath,
                genTaskfilePath,
                activeFileName,
                astJsonPath,
              );
            });
            removeFile(tmpPath);
          } else {
            build(
              filePath,
              astOutputPath,
              extPath,
              activeFileDirPath,
              genSimulationPath,
              genTaskfilePath,
              activeFileName,
              astJsonPath,
            );
          }
        } else {
          vscode.window.showErrorMessage(
            `This file contains error(s). Please fix them before proceeding.`,
          );
        }
      } else {
        vscode.window.showWarningMessage(
          `File extension ${activeFileExt} is NOT supported.`,
        );
      }
    }
  });
  context.subscriptions.push(build_cmd);

  const check_cmd = vscode.commands.registerCommand("Check", () => {
    terminal = terminalSetup(terminal);
    if (astYamlPath != "" && simulationYamlPath != "") {
      terminal?.show();
      terminal?.sendText(`cd ${cdDirPath}`);

      const DSE_REPORT_IMAGE = "ghcr.io/boschglobal/dse-report:latest";
      const simVolumePath = isCodespace
        ? `${cdDirPath}/out/sim`
        : `$(pwd)/out/sim`;

      // Pull the latest report image
      terminal?.sendText(`docker pull ${DSE_REPORT_IMAGE}`);

      // Run docker report command
      const reportCmd = `docker run -it --rm -v ${simVolumePath}:/sim ${DSE_REPORT_IMAGE} report /sim`;

      terminal?.sendText(reportCmd);
    } else {
      vscode.window.showWarningMessage(
        `Please run the DSE build command to Generate the files required for the check command.`,
      );
    }
  });
  context.subscriptions.push(check_cmd);

  const run_cmd = vscode.commands.registerCommand("Run", () => {
    clearTempFiles();
    terminal = terminalSetup(terminal);
    const editor = vscode.window.activeTextEditor;
    if (editor) {
      const [filePath, activeFileExt, activeFileName, activeFileDirPath] =
        getActiveFileInfo(editor);
      if (astYamlPath != "") {
        terminal?.show();
        terminal?.sendText(`cd ${cdDirPath}`);
        //if 'pre_run.sh' is present it gets executed first.
        const execFile = "pre_run.sh";
        const tmpPath: string = path.join(tmpdir(), tmpPreRun);
        const preRunCompletionStatusFile = isCodespace
          ? tmpPath
          : convertToMntPath(tmpPath.replace(/\\/g, "/"));
        const preRunPath = path.join(activeFileDirPath, execFile);
        if (fs.existsSync(preRunPath)) {
          terminal?.sendText(
            `sh ${execFile} && touch ${preRunCompletionStatusFile}`,
          );
          waitForFile(tmpPath, () => {
            console.log(`executing ${execFile}`);
            run(astYamlPath, activeFileDirPath);
          });
          removeFile(tmpPath);
        } else {
          run(astYamlPath, activeFileDirPath);
        }
      } else {
        vscode.window.showWarningMessage(
          `Please run the DSE build command to process dse supported files.`,
        );
      }
    }
  });
  context.subscriptions.push(run_cmd);

  const clean_cmd = vscode.commands.registerCommand("Clean", () => {
    clearTempFiles();
    terminal = terminalSetup(terminal);
    terminal?.show();
    const execFile = "pre_clean.sh";
    const tmpPath: string = path.join(tmpdir(), tmpPreClean);
    const preCleanCompletionStatusFile = isCodespace
      ? tmpPath
      : convertToMntPath(tmpPath.replace(/\\/g, "/"));
    const preCleanPath = path.join(dseDirPath, execFile);
    if (fs.existsSync(preCleanPath)) {
      terminal?.sendText(
        `sh ${execFile} && touch ${preCleanCompletionStatusFile}`,
      );
      waitForFile(tmpPath, () => {
        console.log(`executing ${execFile}`);
        clean(false);
      });
      removeFile(tmpPath);
    } else {
      clean(false);
    }
  });
  context.subscriptions.push(clean_cmd);

  const cleanall_cmd = vscode.commands.registerCommand("Cleanall", () => {
    clearTempFiles();
    terminal = terminalSetup(terminal);
    terminal?.show();
    const execFile = "pre_clean.sh";
    const tmpPath: string = path.join(tmpdir(), tmpPreCleanall);
    const preCleanCompletionStatusFile = isCodespace
      ? tmpPath
      : convertToMntPath(tmpPath.replace(/\\/g, "/"));
    const preCleanPath = path.join(dseDirPath, execFile);
    if (fs.existsSync(preCleanPath)) {
      terminal?.sendText(
        `sh ${execFile} && touch ${preCleanCompletionStatusFile}`,
      );
      waitForFile(tmpPath, () => {
        console.log(`executing ${execFile}`);
        clean(true);
      });
      removeFile(tmpPath);
    } else {
      clean(true);
    }
  });
  context.subscriptions.push(cleanall_cmd);

  const serverModule = context.asAbsolutePath(
    path.join("server", "out", "server.js"),
  );

  // If the extension is launched in debug mode then the debug server options are used
  // Otherwise the run options are used
  const serverOptions: ServerOptions = {
    run: { module: serverModule, transport: TransportKind.ipc },
    debug: {
      module: serverModule,
      transport: TransportKind.ipc,
    },
  };

  // Options to control the language client
  const clientOptions: LanguageClientOptions = {
    documentSelector: [{ scheme: "file", language: "dse" }],
    synchronize: {
      // Notify the server about file changes to '.clientrc files contained in the workspace
      fileEvents: vscode.workspace.createFileSystemWatcher("**/.clientrc"),
    },
  };

  // Create the language client and start the client.
  client = new LanguageClient("dse", "DSE", serverOptions, clientOptions);
  // Start the client. This will also launch the server
  client.start();
}

vscode.window.onDidCloseTerminal((closedTerminal) => {
  if (closedTerminal === terminal) {
    terminal = undefined;
  }
});

function waitForFile(path: string, callback: () => void, interval = 1000) {
  let hasRun = false;
  const checkInterval = setInterval(() => {
    if (fs.existsSync(path) && !hasRun) {
      hasRun = true;
      clearInterval(checkInterval);
      callback();
    }
  }, interval);
}

function validateDiagnostics(
  activeEditor: typeof vscode.window.activeTextEditor,
  diagnosticCollection: vscode.DiagnosticCollection,
): boolean {
  if (!activeEditor) {
    return false;
  }
  const documentUri = activeEditor?.document.uri;
  const existingDiagnostics = diagnosticCollection.get(documentUri);
  if (existingDiagnostics && existingDiagnostics.length > 0) {
    return false;
  }
  return true;
}

function build(
  filePath: string,
  astOutputPath: string,
  extPath: string,
  activeFileDirPath: string,
  genSimulationPath: string,
  genTaskfilePath: string,
  activeFileName: string,
  astJsonPath: string,
) {
  const DSE_BUILDER_IMAGE = "ghcr.io/boschglobal/dse-builder:latest";
  const dockerConfigDir = isCodespace
                        ? "/var/lib/docker/codespacemount/.persistedshare/.docker"
                        : `\${DOCKER_CONFIG:-$HOME/.docker}`;
  const dseScriptName = path.basename(filePath);
  const workdir = activeFileDirPath;

  const tmpPathBuild = path.join(tmpdir(), "build_completed");
  const buildCompletionStatusFile = isCodespace
    ? tmpPathBuild
    : convertToMntPath(tmpPathBuild.replace(/\\/g, "/"));

  const loginCmd = `echo "$AR_TOKEN" | docker login -u "$AR_USER" --password-stdin "$DSE_DOCKER_REPO"`;

  // Get git repo root and project directory (similar to Makefile)
  let repoRoot = workdir;
  let projDir = workdir;
  if (!isCodespace) {
    // Get git repo root: git rev-parse --show-toplevel
    exec("git rev-parse --show-toplevel", { cwd: workdir }, async (err, stdout) => {
      if (!err && stdout) {
        repoRoot = stdout.trim();
      }

      // Get git prefix: git rev-parse --show-prefix
      exec("git rev-parse --show-prefix", { cwd: workdir }, async (prefixErr, prefixStdout) => {
        if (!prefixErr && prefixStdout) {
          const prefix = prefixStdout.trim();
          projDir = `/repo/${prefix}`;
        } else {
          projDir = `/repo`;
        }

        // Build docker command with paths similar to Makefile
        const workdirMnt = convertToMntPath(workdir.replace(/\\/g, "/"));
        const repoRootMnt = convertToMntPath(repoRoot.replace(/\\/g, "/"));
        const envFilePath = await promptForEnvFilePrefix();
        terminal?.show();
        const envFileArg = envFilePath ? `--env-file ${convertToMntPath(envFilePath.replace(/\\/g, "/"))} \\\n          ` : "";

        const dockerCmd = `docker run -it --rm \\
          --user $(id -u):$(id -g) \\
          --group-add $(stat -c '%g' /var/run/docker.sock) \\
          -v ${dockerConfigDir}:/docker-config:ro \\
          -v ${workdirMnt}:/workdir \\
          -v ${repoRootMnt}:/repo \\
          -w ${projDir} \\
          ${envFileArg}-e DOCKER_CONFIG=/docker-config \\
          -e HOME=/workdir \\
          -e PROJDIR=${projDir} \\
          -e WORKDIR=${projDir} \\
          -e ENTRYWORKDIR=${workdirMnt} \\
          -e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \\
          -e DSE_DOCKER_REPO \\
          -v /var/run/docker.sock:/var/run/docker.sock \\
          ${DSE_BUILDER_IMAGE} ${dseScriptName} && touch ${buildCompletionStatusFile}`;

        const cmd = dockerLoginNeeded()
          ? `${loginCmd} && ${dockerCmd}`
          : dockerCmd;
        terminal?.sendText(cmd);
      });
    });
  } else {
    // For Codespace, also calculate git paths for nested Docker containers
    exec("git rev-parse --show-toplevel", { cwd: workdir }, async (err, stdout) => {
      if (!err && stdout) {
        repoRoot = stdout.trim();
      }

      exec("git rev-parse --show-prefix", { cwd: workdir }, async (prefixErr, prefixStdout) => {
        if (!prefixErr && prefixStdout) {
          const prefix = prefixStdout.trim();
          projDir = `/repo/${prefix}`;
        } else {
          projDir = `/repo`;
        }
        // In GitHub Codespaces using docker-outside-of-docker (DooD), the Docker daemon
        // does not resolve the workspace through the dev container's `/workspaces` mount.
        // Instead, it sees the repository under
        // `/var/lib/docker/codespacemount/workspace`. Without this remapping, bind
        // mounts expose an incomplete/stale workspace (e.g. only the `out/` directory),
        // causing the builder to fail to locate files such as `openloop.dse`.
        // Host paths seen by the Docker daemon
        const dockerWorkdir = workdir.replace(
          "/workspaces",
          daemonWorkspace
        );

        const dockerRepoRoot = repoRoot.replace(
          "/workspaces",
          daemonWorkspace
        );
        // In GitHub Codespaces, the Docker socket's group ID as seen inside the
        // builder container can differ from the host's group ID (e.g. host=0,
        // container=800). Determine the GID from the container's perspective so
        // `--group-add` grants the builder access to the mounted Docker socket for
        // nested Docker commands.
        // Determine the socket GID as seen INSIDE the builder container
        const socketGidCmd = `docker run --rm \
          -v /var/run/docker.sock:/var/run/docker.sock \
          --entrypoint sh \
          ${DSE_BUILDER_IMAGE} \
          -c 'stat -c "%g" /var/run/docker.sock'`;

        exec(socketGidCmd, async (gidErr, gidStdout) => {
          const socketGid =
            !gidErr && gidStdout ? gidStdout.trim() : "0";

          console.log("Docker socket GID:", socketGid);
          const envFilePath = await promptForEnvFilePrefix();
          terminal?.show();
          const envFileArg = envFilePath ? `--env-file ${convertToMntPath(envFilePath.replace(/\\/g, "/"))} \\\n            ` : "";

          const dockerCmd = `docker run -it --rm \\
            --user $(id -u):$(id -g) \\
            --group-add ${socketGid} \\
            -v ${dockerConfigDir}:/docker-config:ro \\
            -v ${dockerWorkdir}:/workdir \\
            -v ${daemonWorkspace}:/workspaces \\
            -v ${dockerRepoRoot}:/repo \\
            -w ${projDir} \\
            ${envFileArg}-e DOCKER_CONFIG=/docker-config \\
            -e HOME=/workdir \\
            -e PROJDIR=${projDir} \\
            -e WORKDIR=${projDir} \\
            -e ENTRYWORKDIR=${dockerWorkdir} \\
            -e AR_USER -e AR_TOKEN -e GHE_USER -e GHE_TOKEN -e GHE_PAT \\
            -e DSE_DOCKER_REPO \\
            -v /var/run/docker.sock:/var/run/docker.sock \\
            ${DSE_BUILDER_IMAGE} ${dseScriptName} && touch ${buildCompletionStatusFile}`;

          const cmd = dockerLoginNeeded()
            ? `${loginCmd} && ${dockerCmd}`
            : dockerCmd;
          terminal?.sendText(cmd);
        });
      });
    });
  }

  simulationYamlPath = path.join(activeFileDirPath, "out/simulation.yaml");

  const startTime = Date.now();
  const interval = setInterval(() => {
    if (fs.existsSync(genSimulationPath) && fs.existsSync(genTaskfilePath) && fs.existsSync(tmpPathBuild)) {
      clearInterval(interval);
      // openFile(genSimulationPath); // makes activeFileDirPath to the out folder path
      // removeFile(path.join(activeFileDirPath, 'out', activeFileName + ".json"));      

      tmpterminal?.sendText(`rm -f /tmp/dse_*`);
      setVars(astJsonPath, terminal);

      //if 'post_build.sh' is present in active dsl dir path it gets executed.
      const execFile = "post_build.sh";
      const postBuildPath = path.join(activeFileDirPath, execFile);
      if (fs.existsSync(postBuildPath)) {
        waitForFile(tmpPathBuild, () => {
          console.log(`executing ${execFile}`);
          terminal?.sendText(`sh ${execFile}`);
          removeFile(tmpPathBuild);
        });
      }
    } else if (Date.now() - startTime > timeout) {
      clearInterval(interval);
    }
  }, checkInterval);
}

function run(astYamlPath: string, activeFileDirPath: string) {
  const DSE_SIMER_IMAGE = "ghcr.io/boschglobal/dse-simer:latest";
  const simPath = isCodespace
    ? path.join(activeFileDirPath.replace("/workspaces", daemonWorkspace), "out/sim")
    : convertToMntPath(path.join(activeFileDirPath, "out/sim").replace(/\\/g, "/"));

  const tmpPath = path.join(tmpdir(), tmpSimRun);
  const simCompletionStatusFile = isCodespace
    ? tmpPath
    : convertToMntPath(tmpPath.replace(/\\/g, "/"));
  // Docker command for running simulation
  const dockerCmd = `docker run -it --rm -v ${simPath}:/sim ${DSE_SIMER_IMAGE} && touch ${simCompletionStatusFile}`;
  terminal?.sendText(`docker pull ${DSE_SIMER_IMAGE}`);
  terminal?.sendText(dockerCmd);

  const execFile = "post_run.sh";
  const postRunPath = path.join(activeFileDirPath, execFile);
  if (fs.existsSync(postRunPath)) {
    waitForFile(tmpPath, () => {
      console.log(`executing ${execFile}`);
      terminal?.sendText(`sh ${execFile}`);
    });
    removeFile(tmpPath);
  }
}

function clean(all: boolean = false) {
  const tmpPathClean = path.join(tmpdir(), "clean_completed");
  const cleanCompletionStatusFile = isCodespace
    ? tmpPathClean
    : convertToMntPath(tmpPathClean.replace(/\\/g, "/"));

  const tmpPathCleanall = path.join(tmpdir(), "cleanall_completed");
  const cleanallCompletionStatusFile = isCodespace
    ? tmpPathCleanall
    : convertToMntPath(tmpPathCleanall.replace(/\\/g, "/"));

  const execFile = "post_clean.sh";
  const postCleanPath = path.join(dseDirPath, execFile);

  if (all === false) {
    const cleanCmd = isCodespace
      ? `sudo sh -c 'if [ -d out/ ]; then find out -mindepth 1 -maxdepth 1 ! -name downloads -exec rm -rf {} +; fi' && touch ${cleanCompletionStatusFile}`
      : `if [ -d out/ ]; then find out -mindepth 1 -maxdepth 1 ! -name downloads -exec rm -rf {} +; fi && touch ${cleanCompletionStatusFile}`;
    terminal?.sendText(cleanCmd);
    if (fs.existsSync(postCleanPath)) {
      waitForFile(tmpPathClean, () => {
        console.log(`executing ${execFile}`);
        terminal?.sendText(`sh ${execFile}`);
        removeFile(tmpPathClean);
      });
    }
  } else {
    const cleanallCmd = isCodespace
      ? `sudo rm -rf out && touch ${cleanallCompletionStatusFile}`
      : `rm -rf out && touch ${cleanallCompletionStatusFile}`;
    terminal?.sendText(cleanallCmd);
    if (fs.existsSync(postCleanPath)) {
      waitForFile(tmpPathCleanall, () => {
        console.log(`executing ${execFile}`);
        terminal?.sendText(`sh ${execFile}`);
        removeFile(tmpPathCleanall);
      });
    }
  }
}

function setVars(astJsonPath: string, terminal: vscode.Terminal | undefined) {
  try {
    const rawData = fs.readFileSync(astJsonPath, "utf-8");
    const jsonData = JSON.parse(rawData);

    // setting env vars
    jsonData.children.stacks.forEach((stack: { env_vars: any[] }) => {
      stack.env_vars?.forEach((envVar) => {
        const name = envVar.object.payload.env_var_name.value;
        const value = envVar.object.payload.env_var_value.value;
        envVars[name] = value;
        terminal?.sendText(`export ${name}=${value}`);
      });
    });
  } catch (err) {
    console.error(err);
  }
}

function getActiveFileInfo(
  editor: vscode.TextEditor,
): [string, string, string, string] {
  const filePath = editor.document.uri.fsPath;
  const activeFileExt = path.extname(filePath);
  const activeFileName = path.basename(filePath, path.extname(filePath));
  const activeFileDirPath = path.dirname(filePath);
  dseDirPath = activeFileDirPath;
  return [filePath, activeFileExt, activeFileName, activeFileDirPath];
}

function removeFile(filePath: string) {
  if (fs.existsSync(filePath)) {
    fs.unlinkSync(filePath);
  }
}

function clearTempFiles() {
  const fileNames: string[] = [
    tmpPreBuild,
    tmpPreRun,
    tmpPreClean,
    tmpPreCleanall,
    tmpSimRun,
    tmpBuild,
  ];
  fileNames.forEach((file) => {
    const fullPath = path.join(tmpdir(), file);
    try {
      removeFile(fullPath);
    } catch (err) {
      console.error(err);
    }
  });
}

function terminalSetup(
  terminal: vscode.Terminal | undefined,
): vscode.Terminal | undefined {
  if (!terminal || terminal.exitStatus !== undefined) {
    if (isCodespace) {
      terminal = vscode.window.createTerminal({
        name: "Codespace Terminal",
        shellPath: "/bin/bash",
      });
      console.log("Running inside GitHub Codespaces");
    } else {
      terminal = vscode.window.createTerminal({
        name: "WSL Terminal",
        shellPath: "wsl.exe",
      });
      console.log("Running on local VS Code");
    }
  }
  return terminal;
}

// converting path to the mount path in WSL
function convertToMntPath(winPath: string): string {
  return winPath
    .replace(/\\/g, "/")
    .replace(/^([A-Za-z]):/, (_, drive) => `/mnt/${drive.toLowerCase()}`);
}

function convertToWinPath(mntPath: string): string {
  return mntPath
    .replace(/^\/mnt\/([a-z])\//, (_, drive) => `${drive.toUpperCase()}:\\`)
    .replace(/\//g, "\\");
}

async function promptForEnvFilePrefix(): Promise<string> {
  const loadChoice = await vscode.window.showQuickPick(
    [
      { label: "Yes", description: "Load variables from an env file" },
      { label: "No", description: "Continue without loading an env file" },
    ],
    {
      title: "Build: Load Env File",
      placeHolder: "Select whether to load an env file before running docker build",
      ignoreFocusOut: true,
    },
  );

  if (!loadChoice || loadChoice.label !== "Yes") {
    return "";
  }

  const envCandidateGroups = await Promise.all([
    vscode.workspace.findFiles("**/.env*", "**/node_modules/**", 200),
  ]);
  const envCandidatesMap = new Map<string, vscode.Uri>();
  envCandidateGroups.flat().forEach((uri) => {
    envCandidatesMap.set(uri.fsPath, uri);
  });
  const envCandidates = Array.from(envCandidatesMap.values());

  const envFilePickItems: vscode.QuickPickItem[] = envCandidates.map((uri) => ({
    label: path.basename(uri.fsPath),
    description: uri.fsPath,
  }));

  envFilePickItems.unshift({
    label: "$(folder-opened) Browse...",
    description: "Pick any file from disk",
  });

  const pickedEnvFile = await vscode.window.showQuickPick(envFilePickItems, {
    title: "Build: Select Env File",
    placeHolder: "Choose an env file to pass into the docker build command",
    ignoreFocusOut: true,
  });

  if (!pickedEnvFile) {
    return "";
  }

  let envFilePath = "";
  if (pickedEnvFile.label === "$(folder-opened) Browse...") {
    const selection = await vscode.window.showOpenDialog({
      canSelectMany: false,
      canSelectFiles: true,
      canSelectFolders: false,
      openLabel: "Load Env File",
      filters: {
        "Env Files": ["env*"],
      },
    });

    if (!selection || selection.length === 0) {
      return "";
    }

    envFilePath = selection[0].fsPath;
  } else {
    envFilePath = pickedEnvFile.description ?? "";
  }

  if (!envFilePath) {
    return "";
  }

  const confirmChoice = await vscode.window.showQuickPick(
    [
      { label: "Load", description: path.basename(envFilePath) },
      { label: "Cancel", description: "Do not load env file" },
    ],
    {
      title: "Build: Confirm Env File",
      placeHolder: "Confirm env file selection",
      ignoreFocusOut: true,
    },
  );

  if (!confirmChoice || confirmChoice.label !== "Load") {
    return "";
  }

  vscode.window.showInformationMessage(
    `Using env file: ${path.basename(envFilePath)}`,
  );

  return envFilePath;
}

async function openFile(filePath: string) {
  try {
    const uri = vscode.Uri.file(filePath);
    const document = await vscode.workspace.openTextDocument(uri);
    await vscode.window.showTextDocument(document, { preview: false });
  } catch (error) {
    console.log(`Error opening file: ${error}`);
  }
}

async function dslToAstConvertion(inFilePath: string, extPath: string) {
  const astJsonOutputPath = path.join(extPath, "ast_dag", "ast.json");
  try {
    const dslContent = fs.readFileSync(inFilePath, "utf8");
    const astOutput = parse(dslContent);
    fs.writeFileSync(astJsonOutputPath, JSON.stringify(astOutput), "utf8");
    return true;
  } catch (error) {
    console.error(`exec error: ${error}`);
    return false;
  }
}

async function processAndServeFile(extPath: string): Promise<boolean> {
  updateD3InputFile(extPath);

  // Kill any existing server process
  if (httpServerProcess) {
    httpServerProcess.kill();
    httpServerProcess = null;
  }

  // Clean up lingering processes on current port
  await killProcessOnPort(port);

  // Find an available port (with fallback to basePort)
  const availablePort = await findAvailablePort(basePort);
  port = availablePort;
  console.log(`[INFO] HTTP server will use port ${port}`);

  const fileServePath = path.join(extPath, "ast_dag");
  const file_serve_command = `http-server ${fileServePath} -p ${port} --cors`;

  return new Promise<boolean>((resolve) => {
    try {
      httpServerProcess = exec(file_serve_command, (error, stdout, stderr) => {
        if (error) {
          console.error(`[ERROR] HTTP server failed: ${error.message}`);
          return;
        }
        if (stderr && !stderr.includes("Hit CTRL-C") && !stderr.includes("Started HTTP server")) {
          console.error(`[ERROR] HTTP server stderr: ${stderr}`);
        }
        if (stdout) {
          console.log(`[INFO] HTTP server: ${stdout}`);
        }
      });

      // Handle process exit
      httpServerProcess.on('exit', (code: number) => {
        console.log(`[INFO] HTTP server process exited with code ${code}`);
        httpServerProcess = null;
      });

      // Delay to ensure server starts before checking
      setTimeout(() => {
        waitForHttpServer(port, 40, 250).then(resolve);
      }, 500);
    } catch (err) {
      console.error(`[ERROR] Failed to start HTTP server: ${err}`);
      resolve(false);
    }
  });
}

function isPortAvailable(portNum: number): Promise<boolean> {
  return new Promise((resolve) => {
    const server = require('net').createServer();

    server.once('error', (err: any) => {
      if (err.code === 'EADDRINUSE') {
        resolve(false);
      } else {
        resolve(false);
      }
    });

    server.once('listening', () => {
      server.close();
      resolve(true);
    });

    server.listen(portNum, '0.0.0.0');
  });
}

async function findAvailablePort(startPort: number, maxAttempts: number = 10): Promise<number> {
  let currentPort = startPort;

  for (let i = 0; i < maxAttempts; i++) {
    const available = await isPortAvailable(currentPort);
    if (available) {
      console.log(`[INFO] Port ${currentPort} is available`);
      return currentPort;
    }
    console.log(`[INFO] Port ${currentPort} is in use, trying ${currentPort + 1}`);
    currentPort++;
  }

  console.warn(`[WARN] Could not find available port within ${maxAttempts} attempts, using port ${currentPort}`);
  return currentPort;
}

async function killProcessOnPort(portNum: number): Promise<void> {
  const isWindows = process.platform === 'win32';
  const isLinux = process.platform === 'linux';
  const isMac = process.platform === 'darwin';

  return new Promise((resolve) => {
    if (isWindows) {
      // Windows: use netstat and taskkill
      exec(`netstat -ano | findstr :${portNum}`, (err, stdout) => {
        if (err || !stdout) {
          console.log(`[INFO] No process found on Windows port ${portNum}`);
          resolve();
          return;
        }

        const lines = stdout.split("\n").filter(line => line.includes(`:${portNum}`));
        if (lines.length > 0) {
          const parts = lines[0].trim().split(/\s+/);
          const pid = parts[parts.length - 1];

          if (pid && pid !== 'PID' && !isNaN(parseInt(pid))) {
            console.log(`[INFO] Killing Windows process ${pid} on port ${portNum}`);
            exec(`taskkill /PID ${pid} /F`, () => {
              console.log(`[INFO] Process ${pid} terminated`);
              resolve();
            });
            return;
          }
        }
        resolve();
      });
    } else if (isLinux || isMac) {
      // Linux/macOS: use lsof
      exec(`lsof -i :${portNum} -t 2>/dev/null || true`, (err, stdout) => {
        if (!stdout || stdout.trim() === '') {
          console.log(`[INFO] No process found on Linux/macOS port ${portNum}`);
          resolve();
          return;
        }

        const pids = stdout.trim().split('\n').filter(pid => pid && !isNaN(parseInt(pid)));
        if (pids.length > 0) {
          console.log(`[INFO] Killing Linux/macOS processes on port ${portNum}: ${pids.join(', ')}`);
          const killCmd = `kill -9 ${pids.join(' ')} 2>/dev/null || true`;
          exec(killCmd, () => {
            console.log(`[INFO] Processes terminated`);
            resolve();
          });
          return;
        }
        resolve();
      });
    } else {
      console.log(`[WARN] Unsupported platform for port cleanup`);
      resolve();
    }
  });
}

function waitForHttpServer(
  portNum: number,
  retries: number = 40,
  delayMs: number = 250,
): Promise<boolean> {
  return new Promise((resolve) => {
    const attempt = (retriesLeft: number) => {
      const url = `http://127.0.0.1:${portNum}/input.json`;
      http
        .get(url, (response) => {
          response.resume();
          if (response.statusCode && response.statusCode >= 200 && response.statusCode < 500) {
            console.log(`[INFO] HTTP server on port ${portNum} is ready`);
            resolve(true);
            return;
          }
          if (retriesLeft <= 0) {
            console.error(`[ERROR] HTTP server on port ${portNum} not ready after retries`);
            resolve(false);
            return;
          }
          setTimeout(() => attempt(retriesLeft - 1), delayMs);
        })
        .on("error", (err) => {
          if (retriesLeft <= 0) {
            console.error(`[ERROR] Failed to connect to HTTP server on port ${portNum}: ${err.message}`);
            resolve(false);
            return;
          }
          setTimeout(() => attempt(retriesLeft - 1), delayMs);
        });
    };

    attempt(retries);
  });
}

function generateContainerHTML(
  outputPath: string,
  codespaceHost: string | undefined,
) {
  const url = `https://${codespaceHost}-${port}.app.github.dev/input.json`;
  const htmlContent = `<!DOCTYPE html>
    <html lang="en">
    <head>
        <meta charset="UTF-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1.0" />
        <title>AST</title>
        <style>
            .node {
                text-align: center;
            }
            .link {
                stroke: #00000081;
            }
            .node text {
                font: 14px sans-serif;
                pointer-events: none;
                color: black;
            }
            svg {
                padding: 10px;
            }
        </style>
    </head>
    <body>
        <div class="tree-container">
            <svg></svg>
        </div>
        <script src="https://d3js.org/d3.v6.min.js"></script>
        <script type="text/javascript" src="./ast.js?v=${new Date().getTime()}" codespace_url="${url}"></script>
    </body>
    </html>`;
  fs.writeFileSync(outputPath, htmlContent, "utf8");
}

function getWebviewContent(url: string): string {
  return `
        <!DOCTYPE html>
        <html lang="en">
        <head>
            <meta charset="UTF-8">
            <meta name="viewport" content="width=device-width, initial-scale=1.0">
            <title>Live Preview</title>
            <style>
                html, body {
                    height: 100%;
                    margin: 0;
                    padding: 0;
                    display: flex;
                    justify-content: flex-start;
                    align-items: flex-start;
                    flex-direction: column;
                }
                iframe {
                    width: 100%;
                    height: calc(100% - 40px);
                    border: none;
                    background-color:white;
                    overflow-x: hidden;
                }
            </style>
        </head>
        <body>
            <iframe id="livePreviewIframe" src="${url}?t=${new Date().getTime()}"></iframe>
            <script>
                const vscode = acquireVsCodeApi();

                function refreshIframe() {
                    const iframe = document.getElementById('livePreviewIframe');
                    if (iframe) {
                        iframe.src = "${url}?t=" + new Date().getTime(); // Force reload by appending timestamp
                    }
                }

                // Listen for messages from the extension
                window.addEventListener('message', event => {
                    if (event.data === 'refresh') {
                        refreshIframe();
                    }
                });

                window.addEventListener('resize', function () {
                    const iframe = document.querySelector('iframe');
                    if (iframe) {
                        iframe.style.width = window.innerWidth + 400 + 'px';
                        iframe.style.height = window.innerHeight - 40 + 'px';
                    }
                });

                window.dispatchEvent(new Event('resize'));
            </script>

        </body>
        </html>
    `;
}

function killProcess(portNum: number) {
  // Deprecated: Use killProcessOnPort instead
  killProcessOnPort(portNum).catch(err => {
    console.log(`[INFO] Could not kill process on port ${portNum}`);
  });
}

function jsonFormatterD3(json_data: any): typeof default_struct {
  outJson = { ...default_struct };
  try {
    if (json_data !== undefined) {
      outJson.nodes = [];
      outJson.links = [];

      let model_count = 0;
      for (const stack of json_data.children.stacks) {
        model_count += stack.children.models.length;
      }

      let id = 1;
      for (const stack of json_data.children.stacks) {
        for (const model of stack.children.models) {
          const node_data: Node = {
            id,
            name: model.object.payload.model_name.value,
            type: "rect",
          };
          outJson.nodes.push(node_data);

          for (const channel of model.children.channels) {
            const node_data: Node = {} as Node;
            if (
              !outJson.nodes.find(
                (node) =>
                  node.name === channel.object.payload.channel_name.value,
              )
            ) {
              model_count += 1;
              node_data.id = model_count;
              node_data.name = channel.object.payload.channel_name.value;
              node_data.alias = channel.object.payload.channel_alias.value;
              node_data.type = "vertical_rounded_rect";
              outJson.nodes.push(node_data);
            }
          }
          id += 1;
        }
      }

      for (const channel of json_data.children.channels) {
        const channel_name = channel.object.payload.channel_name.value;
        for (const network of channel.children.networks) {
          const node_data: Node = {} as Node;
          if (
            !outJson.nodes.find(
              (node) => node.name === network.object.payload.network_name.value,
            )
          ) {
            model_count += 1;
            node_data.id = model_count;
            node_data.channel_name = channel_name;
            node_data.name = network.object.payload.network_name.value;
            node_data.mime_type = network.object.payload.mime_type.value;
            node_data.type = "horizontal_rect";
            outJson.nodes.push(node_data);
          }
        }
      }

      for (const stack of json_data.children.stacks) {
        for (const model of stack.children.models) {
          const node_id = outJson.nodes.find(
            (node) => node.name === model.object.payload.model_name.value,
          )!.id;

          for (const channel of model.children.channels) {
            const channel_data = outJson.nodes.find(
              (node) => node.name === channel.object.payload.channel_name.value,
            );
            if (channel_data) {
              const link_data: Link = {
                source: node_id,
                target: channel_data.id,
                type: "link_to_channel",
              };
              outJson.links.push(link_data);
            }

            const channel_name = channel.object.payload.channel_name.value;
            const foundNode = outJson.nodes.find(
              (node) => node.channel_name === channel_name,
            );
            if (foundNode) {
              const can_id = foundNode.id;
              const tmp_link = {
                source: node_id,
                target: can_id,
                type: "link_to_can",
              };
              const exists = outJson.links.some(
                (link) =>
                  link.source === tmp_link.source &&
                  link.target === tmp_link.target,
              );
              exists ? "" : outJson.links.push(tmp_link);
            }
          }
        }
      }

      const targetCount: Record<number, number> = {};
      outJson.links.forEach((link) => {
        targetCount[link.target] = (targetCount[link.target] || 0) + 1;
      });

      for (const tgt in targetCount) {
        outJson.nodes.forEach((node) => {
          if (node.id.toString() === tgt && targetCount[tgt] >= 5) {
            node["type"] = "horizontal_rounded_rect";
          }
        });
      }
    }
  } catch (error) {
    outJson = { ...default_struct };
    console.log(error);
  }
  console.log(JSON.stringify(outJson, null, 2));
  return outJson;
}

function updateD3InputFile(extPath: string): void {
  const file = path.join(extPath, "ast_dag", "ast.json");
  try {
    const data = fs.readFileSync(file, "utf8");
    const json_data = JSON.parse(data);
    const d3Data = jsonFormatterD3(json_data);
    d3Data.nodes.sort((a, b) => a.id - b.id);
    writeFileSync(
      path.join(extPath, "/ast_dag/input.json"),
      JSON.stringify(d3Data, null, 2),
      "utf8",
    );
  } catch (error) {
    console.error("Error preparing D3 input file:", error);
  }
}

function dockerLoginNeeded(): boolean {
  try {
    const dockerConfig = path.join(
      process.env.DOCKER_CONFIG ??
      path.join(process.env.HOME ?? "", ".docker"),
      "config.json",
    );

    const config = JSON.parse(fs.readFileSync(dockerConfig, "utf8"));

    return !config?.auths?.["artifactory.boschdevcloud.com"]?.auth;
  } catch {
    return true;
  }
}

export async function deactivate() {
  console.log(`[INFO] Extension deactivating, cleaning up port ${port}...`);

  if (httpServerProcess) {
    console.log(`[INFO] Killing HTTP server process...`);
    httpServerProcess.kill();
    httpServerProcess = null;
  }

  try {
    await killProcessOnPort(port);
  } catch (err) {
    console.log(`[WARN] Could not cleanup port ${port}: ${err}`);
  }

  if (client) {
    try {
      await client.stop();
    } catch (err) {
      console.log(`[WARN] Error stopping language client: ${err}`);
    }
  }

  console.log(`[INFO] Extension deactivated successfully`);
}
