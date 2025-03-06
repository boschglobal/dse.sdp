import * as path from 'path';
import * as vscode from 'vscode';
import { readFile, watch, writeFileSync } from 'fs';
import { exec } from 'child_process';
import {
    LanguageClient,
    LanguageClientOptions,
    ServerOptions,
    TransportKind
} from 'vscode-languageclient/node';
import * as util from 'util';
import * as fs from 'fs';


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
    links: [] as Link[]
};

let outJson = { ...default_struct };
const port = 3001;
let client: LanguageClient;
const execPromise = util.promisify(exec);
let panel: vscode.WebviewPanel;
let terminal: vscode.Terminal | undefined;
let tmpterminal: vscode.Terminal | undefined;
const supportedExtensions = new Set<string>(['.fsil', '.fs', '.dse']);
const isCodespace = vscode.env.remoteName === "codespaces";
export function activate(context: vscode.ExtensionContext) {
    let activeEditor = vscode.window.activeTextEditor;
    const extPath = vscode.extensions.getExtension('dse.fsil')!.extensionPath;
    const switchPanel = async (isSideBySide: boolean) => {
        activeEditor = vscode.window.activeTextEditor;
        if (activeEditor && activeEditor.document.languageId === 'fsil') {
            const filePath = activeEditor.document.uri.fsPath;
            let convStatus = await dslToAstConvertion(filePath, extPath);
            if (convStatus === true) {
                let status = processAndServeFile(extPath);
                if (status === true) {
                    try {
                        panel.dispose();
                    } catch (error) {
                        console.log(error);
                    }

                    panel = vscode.window.createWebviewPanel(
                        'livePreview',
                        'Live Preview',
                        isSideBySide
                            ? vscode.ViewColumn.Beside  // Open in the side-by-side panel
                            : vscode.ViewColumn.Active, // Open in a single panel
                        {
                            enableScripts: true
                        }
                    );
                    const url = `http://127.0.0.1:${port}/ast.html`;
                    panel.webview.html = getWebviewContent(url);
                    let debounceTimer: NodeJS.Timeout;
                    const debounceDelay = 1000;
                    watch(filePath, async (eventType, filename,) => {
                        if (eventType === 'change') {
                            let status = await dslToAstConvertion(filePath, extPath);
                            if (status === true) {
                                updateD3InputFile(extPath);
                                clearTimeout(debounceTimer);
                                debounceTimer = setTimeout(() => {
                                    const cacheBustedUrl = `${url}?t=${new Date().getTime()}`;
                                    panel.webview.html = getWebviewContent(cacheBustedUrl);
                                }, debounceDelay);
                            }
                        }
                    });
                    vscode.window.showInformationMessage(`Live View created. Listening changes in file ${filePath}`);
                }
            }
        }
    };

    context.subscriptions.push(vscode.commands.registerCommand('livePreview.toggle', () => {
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const filePath = editor.document.uri.fsPath;
            const activeFileExt = path.extname(filePath);
            if (supportedExtensions.has(activeFileExt)) {
                switchPanel(false);  // Open preview in the active panel
            } else {
                vscode.window.showWarningMessage(`File extension ${activeFileExt} is NOT supported.`);
            }
        }
    }));

    context.subscriptions.push(vscode.commands.registerCommand('livePreview.toggleSideBySide', () => {
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const filePath = editor.document.uri.fsPath;
            const activeFileExt = path.extname(filePath);
            if (supportedExtensions.has(activeFileExt)) {
                switchPanel(true);  // Open preview in the side-by-side panel
            } else {
                vscode.window.showWarningMessage(`File extension ${activeFileExt} is NOT supported.`);
            }
        }
    }));

    let build_cmd = vscode.commands.registerCommand('Build', () => {
        terminal = terminalSetup(terminal);
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const [filePath, activeFileExt, activeFileName, activeFileDirPath] = getActiveFileInfo(editor);
            const genSimulationPath = path.join(activeFileDirPath, 'simulation.yaml');
            const genTaskfilePath = path.join(activeFileDirPath, 'Taskfile.yml');
            const astJsonPath = path.join(activeFileDirPath, activeFileName + '.ast.json');
            const astOutputPath = isCodespace ? astJsonPath : convertToMntPath(astJsonPath.replace(/\\/g, "/"));
            if (supportedExtensions.has(activeFileExt)) {
                terminal?.show();
                tmpterminal = terminalSetup(tmpterminal);
                let astYamlPath = path.join(activeFileDirPath, activeFileName + '.yaml');
                astYamlPath = isCodespace ? astYamlPath : convertToMntPath(astYamlPath.replace(/\\/g, "/"));
                removeFile(astYamlPath);
                removeFile(genTaskfilePath);
                removeFile(genSimulationPath);
                terminal?.sendText(`parse2ast ${isCodespace ? filePath : convertToMntPath(filePath.replace(/\\/g, "/"))} ${astOutputPath}`); // executing `parse2ast` command

                const astExecPath = isCodespace ? path.join(extPath, 'bin', 'ast') : convertToMntPath(path.join(extPath, 'bin', 'ast').replace(/\\/g, "/"));
                terminal?.sendText(`${astExecPath} convert -input ${astOutputPath} -output ${astYamlPath}`); // executing ast `convert` command
                terminal?.sendText(`${astExecPath} resolve -input ${astYamlPath} -cache out/cache`); // executing ast `resolve` command

                const genFilesPath = isCodespace ? activeFileDirPath : convertToMntPath(activeFileDirPath.replace(/\\/g, "/"));
                terminal?.sendText(`${astExecPath} generate -input ${astYamlPath} -output ${genFilesPath} -taskfile true`);
                terminal?.sendText(`${astExecPath} generate -input ${astYamlPath} -output ${genFilesPath} -simulation true`);

                tmpterminal?.sendText('touch /tmp/done'); // for checking command completion
                const checkInterval = 1000;
                const timeout = 30000;
                const startTime = Date.now();
                const interval = setInterval(() => {
                    if (fs.existsSync(genSimulationPath) && fs.existsSync(genTaskfilePath)) {
                        clearInterval(interval);
                        openFile(genTaskfilePath);
                        openFile(genSimulationPath);
                        tmpterminal?.sendText('rm /tmp/done');
                        removeFile(astJsonPath);
                        removeFile(astYamlPath);
                    } else if (Date.now() - startTime > timeout) {
                        clearInterval(interval);
                    }
                }, checkInterval);
            } else {
                vscode.window.showWarningMessage(`File extension ${activeFileExt} is NOT supported.`);
            }
        }
    });
    context.subscriptions.push(build_cmd);

    let check_cmd = vscode.commands.registerCommand('Check', () => {
        terminal = terminalSetup(terminal);
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const [filePath, activeFileExt, activeFileName, activeFileDirPath] = getActiveFileInfo(editor);
            vscode.window.showInformationMessage(`Check command called.`);
        }
    });
    context.subscriptions.push(check_cmd);

    let run_cmd = vscode.commands.registerCommand('Run', () => {
        terminal = terminalSetup(terminal);
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const [filePath, activeFileExt, activeFileName, activeFileDirPath] = getActiveFileInfo(editor);
            vscode.window.showInformationMessage(`Run command called.`);
        }
    });
    context.subscriptions.push(run_cmd);

    let clean_cmd = vscode.commands.registerCommand('Clean', () => {
        terminal = terminalSetup(terminal);
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const [filePath, activeFileExt, activeFileName, activeFileDirPath] = getActiveFileInfo(editor);
            vscode.window.showInformationMessage(`Clean command called.`);
        }
    });
    context.subscriptions.push(clean_cmd);

    let cleanall_cmd = vscode.commands.registerCommand('Cleanall', () => {
        terminal = terminalSetup(terminal);
        const editor = vscode.window.activeTextEditor;
        if (editor) {
            const [filePath, activeFileExt, activeFileName, activeFileDirPath] = getActiveFileInfo(editor);
            vscode.window.showInformationMessage(`Cleanall command called.`);
        }
    });
    context.subscriptions.push(cleanall_cmd);


    const serverModule = context.asAbsolutePath(
        path.join('server', 'out', 'server.js')
    );

    // If the extension is launched in debug mode then the debug server options are used
    // Otherwise the run options are used
    const serverOptions: ServerOptions = {
        run: { module: serverModule, transport: TransportKind.ipc },
        debug: {
            module: serverModule,
            transport: TransportKind.ipc,
        }
    };

    // Options to control the language client
    const clientOptions: LanguageClientOptions = {
        documentSelector: [{ scheme: 'file', language: 'fsil' }],
        synchronize: {
            // Notify the server about file changes to '.clientrc files contained in the workspace
            fileEvents: vscode.workspace.createFileSystemWatcher('**/.clientrc')
        }
    };

    // Create the language client and start the client.
    client = new LanguageClient(
        'fsil',
        'Fsil',
        serverOptions,
        clientOptions
    );
    // Start the client. This will also launch the server
    client.start();
}

vscode.window.onDidCloseTerminal((closedTerminal) => {
    if (closedTerminal === terminal) {
        terminal = undefined;
    }
});

function getActiveFileInfo(editor: vscode.TextEditor): [string, string, string, string]{ 
    const filePath = editor.document.uri.fsPath;
    const activeFileExt = path.extname(filePath);
    const activeFileName = path.basename(filePath, path.extname(filePath));
    const activeFileDirPath = path.dirname(filePath);
    return [filePath, activeFileExt, activeFileName, activeFileDirPath]; 
}

function removeFile(filePath: string){
    if (fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
    }
}


function terminalSetup(terminal: vscode.Terminal | undefined): vscode.Terminal | undefined {
    if (!terminal || terminal.exitStatus !== undefined) {
        if (isCodespace) {
            terminal = vscode.window.createTerminal({ name: "Codespace Terminal", shellPath: "/bin/bash" });
            console.log("Running inside GitHub Codespaces");
        } else {
            terminal = vscode.window.createTerminal({ name: "WSL Terminal", shellPath: "wsl.exe" });
            console.log("Running on local VS Code");
        }
    }
    return terminal
}

// converting path to the mount path in WSL
function convertToMntPath(winPath: string): string {
    return winPath.replace(/\\/g, "/")
        .replace(/^([A-Za-z]):/, (_, drive) => `/mnt/${drive.toLowerCase()}`);
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
    const astJsonOutputPath = path.join(extPath, 'ast_dag', 'ast.json');
    const command = `dse-parse2ast "${inFilePath}" "${astJsonOutputPath}"`;
    try {
        const { stdout, stderr } = await execPromise(command);
        if (stderr) {
            vscode.window.showErrorMessage(`An error occurred while DSL convertion - ${stderr}`)
            console.error(`stderr: ${stderr}`);
        }
        console.log(`stdout: ${stdout}`);
        return true;
    } catch (error) {
        vscode.window.showErrorMessage(`An error occurred while DSL convertion - ${error}`)
        console.error(`exec error: ${error}`);
        return false;
    }
}

function processAndServeFile(extPath: string) {
    updateD3InputFile(extPath);
    killProcess(port);
    const fileServePath = path.join(extPath, 'ast_dag');
    const file_serve_command = `http-server ${fileServePath} -p ${port}`;
    exec(file_serve_command, (error, stdout, stderr) => {
        if (error) {
            return false;
        }
        if (stderr) {
            return false;
        }
        console.log(`stdout: ${stdout}`);
    });
    return true;
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
            <iframe id="livePreviewIframe" src="${url}?t=${new Date().getTime()}" sandbox="allow-scripts allow-same-origin"></iframe>
            <script>
                const vscode = acquireVsCodeApi();
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

function killProcess(port: number) {
    try {
        exec(`netstat -ano | findstr :${port}`, (err, stdout, stderr) => {
            if (err) {
                return;
            }
            const lines = stdout.split('\n').filter(line => line.includes(`:${port}`));
            if (lines.length > 0) {
                console.log(lines[0].trim().split(/\s+/));
                const pid = lines[0].trim().split(/\s+/).pop();
                if (pid) {
                    exec(`taskkill /PID ${pid} /F`, (killErr, killStdout, killStderr) => {
                        if (killErr) {
                            return;
                        }
                    });
                }
            }
        });
    } catch (error) {
        console.error(error);
    }
}

function jsonFormatterD3(json_data: any): typeof default_struct {
    outJson = { ...default_struct };
    try {
        if (json_data !== undefined) {
            outJson.nodes = [];
            outJson.links = [];

            let model_count = 0;
            for (let stack of json_data.children.stacks) {
                model_count += stack.children.models.length;
            }

            let id = 1;
            for (let stack of json_data.children.stacks) {
                for (let model of stack.children.models) {
                    let node_data: Node = { id, name: model.object.payload.model_name.value, type: "rect" };
                    outJson.nodes.push(node_data);

                    for (let channel of model.children.channels) {
                        let node_data: Node = {} as Node;
                        if (!outJson.nodes.find(node => node.name === channel.object.payload.channel_name.value)) {
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

            for (let channel of json_data.children.channels) {
                const channel_name = channel.object.payload.channel_name.value;
                for (let network of channel.children.networks) {
                    let node_data: Node = {} as Node;
                    if (!outJson.nodes.find(node => node.name === network.object.payload.network_name.value)) {
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

            for (let stack of json_data.children.stacks) {
                for (let model of stack.children.models) {
                    const node_id = (outJson.nodes.find(node => node.name === model.object.payload.model_name.value))!.id;

                    for (let channel of model.children.channels) {

                        const channel_data = (outJson.nodes.find(node => node.name === channel.object.payload.channel_name.value));
                        if (channel_data) {
                            let link_data: Link = { source: node_id, target: channel_data.id, type: 'link_to_channel' };
                            outJson.links.push(link_data);
                        }


                        const channel_name = channel.object.payload.channel_name.value;
                        const foundNode = outJson.nodes.find(node => node.channel_name === channel_name);
                        if (foundNode) {
                            const can_id = foundNode.id;
                            const tmp_link = { source: node_id, target: can_id, type: 'link_to_can' }
                            const exists = outJson.links.some(link => link.source === tmp_link.source && link.target === tmp_link.target);
                            exists ? "" : outJson.links.push(tmp_link)
                        }
                    }
                }
            }

            const targetCount: Record<number, number> = {};
            outJson.links.forEach(link => {
                targetCount[link.target] = (targetCount[link.target] || 0) + 1;
            });

            for (const tgt in targetCount) {
                outJson.nodes.forEach(node => {
                    if (node.id.toString() === tgt && targetCount[tgt] >= 5) {
                        node["type"] = 'horizontal_rounded_rect';
                    }
                });
            }
        }
    } catch (error) {
        outJson = { ...default_struct };
        console.log(error);
    }
    return outJson;
}


function updateD3InputFile(extPath: string): void {
    const file = path.join(extPath, 'ast_dag', 'ast.json');
    readFile(file, 'utf8', (err, data) => {
        if (err) {
            console.error('Error reading the file:', err);
            return;
        }

        let json_data: any;
        try {
            json_data = JSON.parse(data);
            console.log(json_data);
        } catch (error) {
            console.log('Error parsing JSON:', error);
            return;
        }
        const d3Data = jsonFormatterD3(json_data);
        d3Data.nodes.sort((a, b) => a.id - b.id);
        writeFileSync(path.join(extPath, '/ast_dag/input.json'), JSON.stringify(d3Data, null, 2), 'utf8');
    });
}

export function deactivate(): Thenable<void> | undefined {
    if (!client) {
        return undefined;
    }
    killProcess(port);
    return client.stop();
}
