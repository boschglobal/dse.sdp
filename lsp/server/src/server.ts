import * as fs from 'fs';
import * as path from 'path';
import * as https from 'https';
import { HttpsProxyAgent } from 'https-proxy-agent';
import {
	createConnection,
	TextDocuments,
	Diagnostic,
	DiagnosticSeverity,
	ProposedFeatures,
	InitializeParams,
	DidChangeConfigurationNotification,
	CompletionItem,
	CompletionItemKind,
	TextDocumentPositionParams,
	TextDocumentSyncKind,
	InitializeResult,
	DocumentDiagnosticReportKind,
	type DocumentDiagnosticReport,
	InsertTextFormat,
} from 'vscode-languageserver/node';

import {
	TextDocument
} from 'vscode-languageserver-textdocument';
import * as yaml from 'js-yaml';

const agent = new HttpsProxyAgent('http://localhost:3128');

// Create a connection for the server, using Node's IPC as a transport.
// Also include all preview / proposed LSP features.
const connection = createConnection(ProposedFeatures.all);

let metadata_data: any;
let selectedModel: string | undefined = '';
let selectedWorkflow: string | undefined = '';
let channel_data: any = [];
let channels: string[] = [];
let taskfile_data: { [key: string]: any } = {};
let suggestion_data: { [key: string]: any } = {};
let taskfile_vars_suggestions: string[] = [];
let uses_items: { [key: string]: any } = { 'repos': {}, 'files': {} };
let architectures: any[] = [];
let yamlData: any;
let workflowNames: any;

// Create a simple text document manager.
const documents: TextDocuments<TextDocument> = new TextDocuments(TextDocument);

let hasConfigurationCapability = false;
let hasWorkspaceFolderCapability = false;
let hasDiagnosticRelatedInformationCapability = false;

function fillMissingSuggestionData(suggestion_data: any, taskfile_data: any) {
	for (const repo in taskfile_data) {
		for (const model in suggestion_data) {
			if (model.includes(repo)) {
				const suggestion_workflow_obj = suggestion_data[model]['workflows'];
				const suggestion_workflows = suggestion_workflow_obj.map((workflow: {}) => Object.keys(workflow)[0]);
				for (const workflow of suggestion_workflow_obj) {
					const workflow_name = Object.keys(workflow)[0];
					const taskfile_workflow_obj = taskfile_data[repo]["workflows"];
					const taskfile_workflows = taskfile_workflow_obj.map((workflow: {}) => Object.keys(workflow)[0]);
					if (taskfile_workflows.includes(workflow_name)) {
						// To check missing key values of suggestion workflow.
						const task_workflow_idx = taskfile_workflows.indexOf(workflow_name);
						const suggestion_workflow_idx = suggestion_workflows.indexOf(workflow_name);

						const workflow_keys_from_taskfile = Object.keys(taskfile_workflow_obj[task_workflow_idx][workflow_name]);
						const workflow_keys_from_suggestion = Object.keys(suggestion_workflow_obj[suggestion_workflow_idx][workflow_name])
						const missing_keys_in_suggestion: string[] = workflow_keys_from_taskfile.filter(value => !workflow_keys_from_suggestion.includes(value));
						for (const missing_key of missing_keys_in_suggestion) { // Adding missing key values of each taskfile workflow to suggestion workflow.
							suggestion_data[model]['workflows'][suggestion_workflow_idx][workflow_name][missing_key] = taskfile_workflow_obj[task_workflow_idx][workflow_name][missing_key];
						}
						// To check for missing taskfile vars.
						// const taskfile_vars:string[] = taskfile_workflow_obj[task_workflow_idx][workflow_name]['vars'];
						// const suggestion_vars:string[] = suggestion_workflow_obj[suggestion_workflow_idx][workflow_name]['vars'];
						// const missing_vars_in_suggestion : string[] = taskfile_vars.filter(value => !suggestion_vars.includes(value));
					}
				}
			}
		}
	}
}

function metadataDataParser(metadata_data: any) {
	function getWorkflowDetails(workflows: string[]): any[] {
		let ret_workflows: any = [];
		for (let workflow_name of workflows) {
			let workflow_obj: any = {};
			let vars: any = [];
			let required_vars: any = [];
			let vars_desc: any = {};
			let default_values: any = {};
			if (workflow_name in metadata_data['metadata']['tasks']) {
				for (const var_val in metadata_data['metadata']['tasks'][workflow_name]['vars']) {
					const var_object = metadata_data['metadata']['tasks'][workflow_name]['vars'][var_val];
					if (var_object['required'] === true) {
						required_vars.push(var_val);
					}
					else {
						vars.push(var_val);
					}

					if ('hint' in var_object) {
						if (var_object['hint'] != null) {
							vars_desc[var_val] = var_object['hint'];
						}
					}

					if ('default' in var_object) {
						if (var_object['default'] != null) {
							default_values[var_val] = var_object['default'];
						}
					}
				}
				workflow_obj[workflow_name] = {
					'vars': vars,
					'required_vars': required_vars,
					'vars_desc': vars_desc,
					'default_values': default_values,
					'internal': false // default value. later will be updated based on taskfile 'internal' key value in 'fillMissingSuggestionData' function.
				}
				ret_workflows.push(workflow_obj);
			}
			else {
				workflow_obj[workflow_name] = {
					'vars': [],
					'required_vars': [],
					'vars_desc': {},
					'default_values': {},
					'internal': false // default value. later will be updated based on taskfile 'internal' key value in 'fillMissingSuggestionData' function.
				}
				ret_workflows.push(workflow_obj);
			}
		}
		return ret_workflows;
	}

	function getChannels(channels: []): String[] {
		let ret_channels: any = [];
		for (let obj of channels) {
			ret_channels.push(obj['alias'])
		}
		return ret_channels;
	}

	for (const model in metadata_data['metadata']['models']) {
		const model_obj: any = metadata_data['metadata']['models'][model];
		const display_name = model_obj['displayName'];
		const path = model_obj['path'];
		const name = model_obj['name'];
		const workflows = getWorkflowDetails(model_obj['workflows']);
		const platforms = model_obj['platforms'];
		const channels = getChannels(model_obj['channels']);
		suggestion_data[display_name] = {
			'workflows': workflows,
			'path': path,
			'name': name,
			'platforms': platforms,
			'channels': channels
		};
	}

	fillMissingSuggestionData(suggestion_data, taskfile_data);
	const jsonData = JSON.stringify(suggestion_data, null, 2);
	const outputPath = path.join(__dirname, 'suggestion_data.json');
	fs.writeFileSync(outputPath, jsonData, 'utf-8');
}

function taskFileParser(yamlData: any, model: any) {
	if ('vars' in yamlData) {
		if ('global_vars' in taskfile_data[model]) {
			let g_vars = taskfile_data[model]['global_vars'];
			for (const var_name of Object.keys(yamlData.vars)) {
				if (!g_vars.includes(var_name)) {
					g_vars.push(var_name);
				}
			}
			taskfile_data[model]['global_vars'] = g_vars;
		} else {
			taskfile_data[model]['global_vars'] = Object.keys(yamlData.vars);
		}
	}
	taskfile_data[model]['workflows'] = [];
	for (let task of Object.keys(yamlData.tasks)) {
		let item = {};
		let internal = false;
		let required: any[] = [];
		if ('vars' in yamlData.tasks[task]) {
			let vars = Object.keys(yamlData.tasks[task]['vars']);

			if ('internal' in yamlData.tasks[task]) {
				internal = yamlData.tasks[task]['internal'];
			}

			if ('requires' in yamlData.tasks[task]) {
				required = yamlData.tasks[task]['requires']['vars'];
			}

			vars = vars.filter(value => !required.includes(value)); // removing required vars from vars list
			item = {
				'vars': vars,
				'internal': internal,
				'required_vars': required
			};
			taskfile_data[model]['workflows'].push({ [task]: item });
		}
	}
}

function parseTaskfile(yamlData: any, repo: any) {
	taskfile_data[repo] = {};
	taskFileParser(yamlData, repo);
	metadataDataParser(metadata_data);
	// const jsonData = JSON.stringify(taskfile_data, null, 2);
	// const outputPath = path.join(__dirname, 'taskfile_vars.json');
	// fs.writeFileSync(outputPath, jsonData, 'utf-8');
}

function gen_git_raw_url(repo: { [key: string]: any }, file: string): string {
	const pattern: RegExp = /https\:\/\/github\.com\/(\w+)\/(\w+(?:\.\w+))(\/.*)?/;
	let git_link = repo['link'];
	const matchResult = git_link.match(pattern);
	let owner = '';
	let repo_name = '';
	let path = '';
	if (matchResult) {
		owner = matchResult[1];
		repo_name = matchResult[2];
		try {
			path = matchResult[3];
		} catch {
			path = '';
		}
	} else {
		console.log("No match found");
	}

	let raw_url: string = '';
	if (path != undefined) {
		raw_url = `https://raw.githubusercontent.com/${owner}/${repo_name}${path.replace('blob', '')}`;
	} else {
		raw_url = `https://raw.githubusercontent.com/${owner}/${repo_name}/${repo['version']}/${file}`;
	}
	return raw_url;
}

async function fetchGitHubRawFile(url: string) {
	return new Promise((resolve, reject) => {
		https.get(url, { agent }, (res) => {
			if (res.statusCode === 404) {
				res.on('end', () => {
					resolve('404');
				});
			}
			let data = '';
			res.on('data', (chunk) => {
				data += chunk;
			});

			res.on('end', () => {
				resolve(data);
			});

			res.on('error', (error) => {
				reject(error);
			});
		});
	});
}

connection.onInitialize((params: InitializeParams) => {
	const capabilities = params.capabilities;

	hasConfigurationCapability = !!(
		capabilities.workspace && !!capabilities.workspace.configuration
	);
	hasWorkspaceFolderCapability = !!(
		capabilities.workspace && !!capabilities.workspace.workspaceFolders
	);
	hasDiagnosticRelatedInformationCapability = !!(
		capabilities.textDocument &&
		capabilities.textDocument.publishDiagnostics &&
		capabilities.textDocument.publishDiagnostics.relatedInformation
	);

	const result: InitializeResult = {
		capabilities: {
			textDocumentSync: 1, //TextDocumentSyncKind.Incremental,
			completionProvider: {
				resolveProvider: true,
				triggerCharacters: ['='] // Include '=' as a trigger character
			},
			diagnosticProvider: {
				interFileDependencies: false,
				workspaceDiagnostics: false
			}
		}
	};
	if (hasWorkspaceFolderCapability) {
		result.capabilities.workspace = {
			workspaceFolders: {
				supported: true
			}
		};
	}
	return result;
});

connection.onInitialized(() => {
	if (hasConfigurationCapability) {
		// Register for all configuration changes.
		connection.client.register(DidChangeConfigurationNotification.type, undefined);
	}
	if (hasWorkspaceFolderCapability) {
		connection.workspace.onDidChangeWorkspaceFolders(_event => {
			connection.console.log('Workspace folder change event received.');
		});
	}
});

interface ExampleSettings {
	maxNumberOfProblems: number;
}

const defaultSettings: ExampleSettings = { maxNumberOfProblems: 1000 };
let globalSettings: ExampleSettings = defaultSettings;

// Cache the settings of all open documents
const documentSettings: Map<string, Thenable<ExampleSettings>> = new Map();

connection.onDidChangeConfiguration(change => {
	if (hasConfigurationCapability) {
		// Reset all cached document settings
		documentSettings.clear();
	} else {
		globalSettings = <ExampleSettings>(
			(change.settings.fsil || defaultSettings)
		);
	}
	connection.languages.diagnostics.refresh();
});

function getDocumentSettings(resource: string): Thenable<ExampleSettings> {
	if (!hasConfigurationCapability) {
		return Promise.resolve(globalSettings);
	}
	let result = documentSettings.get(resource);
	if (!result) {
		result = connection.workspace.getConfiguration({
			scopeUri: resource,
			section: 'fsil'
		});
		documentSettings.set(resource, result);
	}
	return result;
}

// Only keep settings for open documents
documents.onDidClose(e => {
	documentSettings.delete(e.document.uri);
});

connection.languages.diagnostics.on(async (params) => {
	const document = documents.get(params.textDocument.uri);
	if (document !== undefined) {
		return {
			kind: DocumentDiagnosticReportKind.Full,
			items: await validateTextDocument(document)
		} satisfies DocumentDiagnosticReport;
	} else {
		return {
			kind: DocumentDiagnosticReportKind.Full,
			items: []
		} satisfies DocumentDiagnosticReport;
	}
});

function fetchGitData(uses_items: { [key: string]: any }) {
	for (let repo in uses_items['repos']) {
		const metadata_git_raw_url: string = gen_git_raw_url(uses_items['repos'][repo], 'Metadata.yml');
		const taskfile_git_raw_url: string = gen_git_raw_url(uses_items['repos'][repo], 'Taskfile.yml');

		fetchGitHubRawFile(metadata_git_raw_url).then((content: any) => {
			if (content != undefined) {
				if (content === '404') {
					return ''; // no metadata file, ie no models to display.
				} else {
					metadata_data = yaml.load(content);
					fetchGitHubRawFile(taskfile_git_raw_url).then((content: any) => {
						if (content != undefined) {
							yamlData = yaml.load(content);
							parseTaskfile(yamlData, repo)
						}
					})
						.catch((error) => {
							console.log("Error in fetching taskfile ========>")
							console.error(error);
						});
				}
			}
		}).catch((error) => {
			console.log("Error in fetching metadata file ========>")
			console.error(error);
		});
	}
}

documents.onDidChangeContent(change => {
	validateTextDocument(change.document);
	getUsesItems(change.document);
	getSelectedModelName(change.document);
});

async function getSelectedModelName(textDocument: TextDocument) {
	try {
		channels = [];
		const text = textDocument.getText();
		let modelMatch: RegExpExecArray | null;
		let matches: string[] = [];
		const modelPattern = /\b^model\s*\w+\s+(\S+)\s+.*\b/gm
		while ((modelMatch = modelPattern.exec(text)) !== null) {
			matches.push(modelMatch[1]);
		}
		selectedModel = matches[matches.length - 1];
		console.log('selectedModel is : ', selectedModel);
		if (selectedModel !== undefined && selectedModel !== '') {
			workflowNames = suggestion_data[selectedModel]["workflows"].map((workflow: {}) => Object.keys(workflow)[0]);
		}
	} catch {
		return [];
	}
}

async function getUsesItems(textDocument: TextDocument) {
	uses_items = { 'repos': {}, 'files': {} };
	const text = textDocument.getText();
	const uses_pattern = /\b(uses)\b/g
	const uses_item_pattern = /\b^(\w+(?:.\w+)?)\s+(https\:\/\/.*)\s(v[.0-9]+)\s*\b/
	let usesMatch: RegExpExecArray | null;
	let usesFlag = false;

	const lines = text.split('\n');
	for (let i = 0; i < lines.length; i++) {
		let line = lines[i];

		usesMatch = uses_pattern.exec(line);
		if (usesMatch) { usesFlag = true; continue; }

		if (usesFlag && line.trim() != "") {
			let uses_items_match = uses_item_pattern.exec(line)
			if (uses_items_match) {
				let repo = uses_items_match[1];
				let git_link = uses_items_match[2];
				let version = uses_items_match[3];

				let link_pattern = /https\:\/\/github\.com\/(\w+)\/(\w+(?:\.\w+))$/;
				let link_match = link_pattern.exec(git_link)
				if (link_match) {
					uses_items['repos'][repo] = {
						'link': git_link,
						'version': version
					};
				} else {
					uses_items['files'][repo] = {
						'link': git_link,
						'version': version
					};
				}
			}
		}
		else {
			usesFlag = false;
		}
	}
	fetchGitData(uses_items);
}

async function validateTextDocument(textDocument: TextDocument): Promise<Diagnostic[]> {
	const settings = await getDocumentSettings(textDocument.uri);
	let problems = 0;
	const diagnostics: Diagnostic[] = [];
	return diagnostics;
}

connection.onDidChangeWatchedFiles(_change => {
	connection.console.log('We received a file change event');
});

// This handler provides the initial list of the completion items.
connection.onCompletion(
	(textDocumentPosition: TextDocumentPositionParams): CompletionItem[] => {
		const completionItems: CompletionItem[] = [];
		const document = documents.get(textDocumentPosition.textDocument.uri);
		if (document) {
			const text = document.getText();
			const lines = text.split('\n');
			const line = lines[textDocumentPosition.position.line];
			const word = line.substring(0, textDocumentPosition.position.character);
			if (selectedModel !== undefined && selectedModel !== '') {
				architectures = suggestion_data[`${selectedModel}`]["platforms"];
				channels = suggestion_data[`${selectedModel}`]["channels"];
			}
			if (word.startsWith('s')) {
				completionItems.length = 0;
				const simulationCompletionItem: CompletionItem = {
					label: "simulation",
					kind: CompletionItemKind.Keyword,
					data: "simulation_keyword"
				};
				const stackCompletionItem: CompletionItem = {
					label: "stack",
					kind: CompletionItemKind.Snippet,
					insertText: "stack\nmodel\n${1:model_name} ${2|" + Object.keys(suggestion_data).join(',') + "|}",
					documentation: 'Inserts a stack snippet',
					insertTextFormat: InsertTextFormat.Snippet,
				};
				completionItems.push(simulationCompletionItem, stackCompletionItem);
			}
			else if (word.startsWith('c')) {
				completionItems.length = 0;
				const completionItem: CompletionItem = {
					label: "channel",
					kind: CompletionItemKind.Keyword,
					data: "channel_keyword"
				};
				completionItems.push(completionItem);
			}
			else if (word.startsWith('w')) {
				completionItems.length = 0;
				const completionItem: CompletionItem = {
					label: "workflow",
					kind: CompletionItemKind.Keyword,
					data: "workflow_keyword"

				};
				completionItems.push(completionItem);
			}
			else if (word.startsWith('m')) {
				completionItems.length = 0;
				const modelCompletionItem: CompletionItem = {
					label: "model",
					kind: CompletionItemKind.Snippet,
					insertText: 'model\n${1:model_name} ${2|' + Object.keys(suggestion_data).join(',') + '|}',
					documentation: 'Inserts a model snippet',
					data: "model_suggestion",
					insertTextFormat: InsertTextFormat.Snippet
				};
				completionItems.push(modelCompletionItem);
			}
			else if (word.startsWith('e')) {
				completionItems.length = 0;
				const completionItem: CompletionItem = {
					label: "env ",
					kind: CompletionItemKind.Keyword,
					data: "env_keyword"
				};
				completionItems.push(completionItem);
			}
			else if (word.startsWith('f')) {
				completionItems.length = 0;
				const completionItem: CompletionItem = {
					label: "from",
					kind: CompletionItemKind.Keyword,
					data: "from_keyword"
				};
				completionItems.push(completionItem);
			}
			else if (word.startsWith('a')) {
				completionItems.length = 0;
				const asCompletionItem: CompletionItem = {
					label: "as ",
					kind: CompletionItemKind.Keyword,
					data: "as_keyword"
				};
				completionItems.push(asCompletionItem);
			}
			else if (word.startsWith('u')) {
				completionItems.length = 0;
				const completionItem: CompletionItem = {
					label: "uses ",
					kind: CompletionItemKind.Keyword,
					data: "uses_keyword"
				};
				completionItems.push(completionItem);
			}
			else if (word.startsWith('n')) {
				completionItems.length = 0;
				const completionItem: CompletionItem = {
					label: "network ",
					kind: CompletionItemKind.Keyword,
					data: "network_keyword"
				};
				completionItems.push(completionItem);
			}
			else if (word.startsWith('v')) {
				completionItems.length = 0;
				const completionItem: CompletionItem = {
					label: "var",
					kind: CompletionItemKind.Keyword,
					data: "var_keyword"
				};
				completionItems.push(completionItem);
			}

			if (word.endsWith('channel')) {
				completionItems.length = 0;
				channels.forEach((channel: string) => {
					const completionItem: CompletionItem = {
						label: channel,
						kind: CompletionItemKind.Value,
						detail: 'channel name',
						filterText: "channel",
						data: "channel_suggestion"
					};
					completionItems.push(completionItem);
				});
			}
			else if (word.endsWith('workflow')) {
				completionItems.length = 0;
				workflowNames.forEach((workflow: string) => {
					const completionItem: CompletionItem = {
						label: workflow,
						kind: CompletionItemKind.Value,
						detail: 'workflow name',
						filterText: "workflow",
						data: "workflow_suggestion"
					};
					completionItems.push(completionItem);
				});
			}
			else if (word.endsWith('from')) {
				completionItems.length = 0;
				Object.keys(uses_items['files']).forEach((file: string) => {
					const completionItem: CompletionItem = {
						label: file,
						kind: CompletionItemKind.Value,
						detail: 'file name',
						filterText: "from",
						data: "from_suggestion"
					};
					completionItems.push(completionItem);
				});
			}
			else if (word.endsWith('arch=')) {
				completionItems.length = 0;
				architectures.forEach((arch: string) => {
					const completionItem: CompletionItem = {
						label: arch,
						kind: CompletionItemKind.Value,
						detail: 'architecture name',
						filterText: "arch=",
						data: "architecture_suggestion"
					};
					completionItems.push(completionItem);
				});
			}
			else if (word.endsWith('var')) {
				completionItems.length = 0;
				taskfile_vars_suggestions.forEach((task: string) => {
					const completionItem: CompletionItem = {
						label: task,
						kind: CompletionItemKind.Value,
						detail: 'var name',
						filterText: "var",
						data: "var_suggestion"
					};
					completionItems.push(completionItem);
				});
			}
		}

		return completionItems;
	}
);

connection.onCompletionResolve(
	(item: CompletionItem): CompletionItem => {
		if (item.data == "channel_keyword") {
			item.label = `${item.label}`;
			item.insertText = `channel`;
			item.command = {
				title: "channels",
				command: "editor.action.triggerSuggest",
			}
		}
		else if (item.data == "channel_suggestion") {
			item.label = `${item.label}`;
			item.insertTextFormat = InsertTextFormat.Snippet;
			item.insertText = `channel ${item.label}`;
		}
		else if (item.data == "workflow_keyword") {
			item.label = `${item.label}`;
			item.insertText = `workflow`;
			item.command = {
				title: "workflows",
				command: "editor.action.triggerSuggest",
			}
		}
		else if (item.data == "workflow_suggestion") {
			taskfile_vars_suggestions = [];
			item.label = `${item.label}`;
			let required_vars = '';
			selectedWorkflow = item.label;
			console.log("selected workflow is : ", selectedWorkflow);
			let index = workflowNames.indexOf(selectedWorkflow);
			const workflow_obj = suggestion_data[`${selectedModel}`]["workflows"][index][`${selectedWorkflow}`]
			if (workflow_obj['internal'] === false) {
				taskfile_vars_suggestions = workflow_obj['vars'];
				const requiredVars = workflow_obj['required_vars'];
				taskfile_vars_suggestions = taskfile_vars_suggestions.filter(item => !requiredVars.includes(item));

				for (let i = 0; i < requiredVars.length; i++) {
					const r_var = requiredVars[i];

					if (r_var in workflow_obj['default_values']) {
						const value_suggestions = [workflow_obj['default_values'][r_var], ...Object.keys(uses_items['files'])];
						required_vars += "\nvar " + r_var + " ${" + (i + 1) + "|" + value_suggestions.join(',') + "|}";
					} else {
						required_vars += "\nvar " + r_var + " ${" + (i + 1) + "|" + Object.keys(uses_items['files']).join(',') + "|}";
					}
				}
			} else {
				taskfile_vars_suggestions = [];
			}
			item.insertTextFormat = InsertTextFormat.Snippet;
			item.insertText = `workflow ${selectedWorkflow}${required_vars}`;
		}
		else if (item.data == "from_keyword") {
			item.label = `${item.label}`;
			item.insertText = `from`;
			item.command = {
				title: "files",
				command: "editor.action.triggerSuggest",
			}
		}
		else if (item.data == "from_suggestion") {
			item.label = `${item.label}`;
			item.insertText = `from ${item.label}`;
		}
		else if (item.data == "var_keyword") {
			item.label = `${item.label}`;
			item.insertText = `var`;
			item.command = {
				title: "vars",
				command: "editor.action.triggerSuggest",
			}
		}
		else if (item.data == "var_suggestion") {
			item.label = `${item.label}`;
			item.insertTextFormat = InsertTextFormat.Snippet;

			const workflow_obj = suggestion_data[`${selectedModel}`]["workflows"][workflowNames.indexOf(selectedWorkflow)][`${selectedWorkflow}`];
			if (item.label in workflow_obj['default_values']) {
				const value_suggestions = Object.keys(uses_items['files']);
				value_suggestions.unshift(workflow_obj['default_values'][item.label]);
				item.insertText = "var " + item.label + " ${1|" + value_suggestions.join(',') + "|}";
			} else {
				item.insertText = "var " + item.label + " ${1|" + Object.keys(uses_items['files']).join(',') + "|}";
			}
			item.detail = workflow_obj['vars_desc'][`${item.label}`];
		}
		return item;
	}
);

documents.listen(connection);
connection.listen();
