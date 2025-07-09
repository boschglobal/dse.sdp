// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import { Lexer, createToken, IToken, ILexingError } from "chevrotain";

const defaultArch = "linux-amd64";
const defaultStepsize = "0.0005";
const defaultEndtime = "0.005";
let parsedStackArch = "";
interface CustomRegExpExecArray extends RegExpExecArray {
  payload?: any;
}

function matchSimulation(text: string) {
  const simulationPattern =
    /^simulation([ ]+arch\=\S+)?([ ]+stepsize\=(?:\d*\.?\d+))?([ ]+endtime\=(?:\d*\.?\d+))?\s*$/;
  const execResult = simulationPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    let simulationArch = defaultArch;
    if (execResult[1] !== undefined) {
      simulationArch = execResult[1];
    }
    let stepsize = "";
    if (execResult[2] !== undefined) {
      stepsize = execResult[2];
    } else {
      stepsize = defaultStepsize;
    }
    let endtime = "";
    if (execResult[3] !== undefined) {
      endtime = execResult[3];
    } else {
      endtime = defaultEndtime;
    }

    const simulationArchStart = execResult.index + "simulation".length + 1;
    const simulationArchEnd = simulationArchStart + simulationArch.length;

    let stepsizeStart = null;
    let stepsizeEnd = null;
    if (stepsize !== "") {
      stepsizeStart =
        execResult.index + "simulation".length + simulationArch.length + 1;
      stepsizeEnd = stepsizeStart + stepsize.length;
    }

    let endtimeStart = null;
    let endtimeEnd = null;
    if (endtime !== "") {
      endtimeStart =
        execResult.index +
        "simulation".length +
        simulationArch.length +
        stepsize.length +
        1;
      endtimeEnd = endtimeStart + endtime.length;
    }

    execResult.payload = {
      simulation_arch: {
        value: simulationArch.replace("arch=", "").trim(),
        token_type: "simulation_arch",
        start_offset: simulationArchStart,
        end_offset: simulationArchEnd,
      },
      stepsize: {
        value: stepsize.replace("stepsize=", "").trim(),
        token_type: "stepsize",
        start_offset: stepsizeStart,
        end_offset: stepsizeEnd,
      },
      endtime: {
        value: endtime.replace("endtime=", "").trim(),
        token_type: "endtime",
        start_offset: endtimeStart,
        end_offset: endtimeEnd,
      },
    };
  }
  return execResult;
}

export const Simulation = createToken({
  name: "Simulation",
  pattern: matchSimulation,
  line_breaks: false,
});

function matchChannel(text: string) {
  const channelPattern = /^channel([ ]+\w+)(?:([ ]+\w+))?\s*$/;
  const execResult = channelPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const channelName = execResult[1];
    let channelAlias = "";
    if (execResult[2] !== undefined) {
      channelAlias = execResult[2];
    }
    const channelNameStart = execResult.index + "channel".length + 1;
    const channelNameEnd = channelNameStart + channelName.length;
    let channelAliasStart = null;
    let channelAliasEnd = null;
    if (channelAlias !== "") {
      channelAliasStart =
        execResult.index + "channel".length + 1 + channelName.length + 1;
      channelAliasEnd = channelAliasStart + channelAlias.length;
    }
    execResult.payload = {
      channel_name: {
        value: channelName.trim(),
        token_type: "channel_name",
        start_offset: channelNameStart,
        end_offset: channelNameEnd,
      },
      channel_alias: {
        value: channelAlias.trim(),
        token_type: "channel_alias",
        start_offset: channelAliasStart,
        end_offset: channelAliasEnd,
      },
    };
  }
  return execResult;
}

export const Channel = createToken({
  name: "Channel",
  pattern: matchChannel,
  line_breaks: false,
});

function matchFile(text: string) {
  const filePattern = /^file([ ]+\S+)([ ]+(?:uses))?([ ]+\S+)?\s*$/;
  const execResult = filePattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const fileName = execResult[1];
    let fileReferenceType = "";
    if (execResult[2] !== undefined) {
      fileReferenceType = execResult[2];
    }
    let fileValue = "";
    if (execResult[3] !== undefined) {
      fileValue = execResult[3];
    }
    const fileNameStart = execResult.index + "file".length + 1;
    const fileNameEnd = fileNameStart + fileName.length;
    let fileReferenceTypeStart = null;
    let fileReferenceTypeEnd = null;
    if (fileReferenceType !== "") {
      fileReferenceTypeStart =
        execResult.index + "file".length + fileName.length + 1;
      fileReferenceTypeEnd = fileReferenceTypeStart + fileReferenceType.length;
    }
    let fileValueStart = null;
    let fileValueEnd = null;
    if (fileValue !== "") {
      fileValueStart =
        execResult.index +
        "file".length +
        fileName.length +
        fileReferenceType.length +
        1;
      fileValueEnd = fileValueStart + fileValue.length;
    }

    execResult.payload = {
      file_name: {
        value: fileName.trim(),
        token_type: "file_name",
        start_offset: fileNameStart,
        end_offset: fileNameEnd,
      },
      file_reference_type: {
        value: fileReferenceType.trim(),
        token_type: "file_reference_type",
        start_offset: fileReferenceTypeStart,
        end_offset: fileReferenceTypeEnd,
      },
      file_value: {
        value: fileValue.trim(),
        token_type: "file_value",
        start_offset: fileValueStart,
        end_offset: fileValueEnd,
      },
    };
  }
  return execResult;
}

export const File = createToken({
  name: "File",
  pattern: matchFile,
  line_breaks: false,
});

function matchNetwork(text: string) {
  const networkPattern = /^network([ ]+\S+)([ ]+\'\S+\')\s*$/;
  const execResult = networkPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const networkName = execResult[1];
    const mimeType = execResult[2];
    const networkNameStart = execResult.index + "network".length + 1;
    const networkNameEnd = networkNameStart + networkName.length;
    const mimeTypeStart =
      execResult.index + "network".length + 1 + networkName.length + 1;
    const mimeTypeEnd = mimeTypeStart + mimeType.length;
    execResult.payload = {
      network_name: {
        value: networkName.trim(),
        token_type: "network_name",
        start_offset: networkNameStart,
        end_offset: networkNameEnd,
      },
      mime_type: {
        value: mimeType.replace(/\'/g, "").trim(),
        token_type: "mime_type",
        start_offset: mimeTypeStart,
        end_offset: mimeTypeEnd,
      },
    };
  }
  return execResult;
}

export const Network = createToken({
  name: "Network",
  pattern: matchNetwork,
  line_breaks: false,
});

function matchUsesKeyword(text: string) {
  const usesKeywordPattern = /^uses[ \t]*$/;
  const execResult = usesKeywordPattern.exec(text) as CustomRegExpExecArray;
  return execResult;
}

export const Uses = createToken({
  name: "Uses",
  pattern: matchUsesKeyword,
  line_breaks: false,
});

function matchUseItem(text: string) {
  const useItemPattern =
    /^(\S+)([ ]+(?:(?:https\:\/\/\S+)|(?:\S+\.\S+)|(?:\S+\/\S+)))([ ]+v\d+(?:\.\d+)*)?(?:[ ]+(path\=\S+))?(?:[ ]+(user\=\S+))?(?:[ ]+(token\=\S+))?\s*$/;
  const execResult = useItemPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const useItem = execResult[1];
    const link = execResult[2];
    let version = "";
    if (execResult[3] !== undefined) {
      version = execResult[3];
    }
    let path = "";
    if (execResult[4] !== undefined) {
      path = execResult[4];
    }
    let user = "";
    if (execResult[5] !== undefined) {
      user = execResult[5];
    }
    let token = "";
    if (execResult[6] !== undefined) {
      token = execResult[6];
    }
    const useItemStart = execResult.index + 1;
    const useItemEnd = useItemStart + useItem.length;
    const linkStart = execResult.index + useItem.length + 1;
    const linkEnd = linkStart + link.length;
    const versionStart = execResult.index + useItem.length + link.length + 1;
    const versionEnd = versionStart + version.length;
    let pathStart = null;
    let pathEnd = null;
    if (path !== "") {
      pathStart =
        execResult.index + useItem.length + link.length + version.length + 1;
      pathEnd = pathStart + path.length;
    }
    let userStart = null;
    let userEnd = null;
    if (user !== "") {
      userStart =
        execResult.index + useItem.length + link.length + version.length + 1;
      userEnd = userStart + user.length;
    }
    let tokenStart = null;
    let tokenEnd = null;
    if (token !== "") {
      tokenStart =
        execResult.index +
        useItem.length +
        link.length +
        version.length +
        user.length +
        1;
      tokenEnd = tokenStart + token.length;
    }
    execResult.payload = {
      use_item: {
        value: useItem.trim(),
        token_type: "use_item",
        start_offset: useItemStart,
        end_offset: useItemEnd,
      },
      link: {
        value: link.trim(),
        token_type: "link",
        start_offset: linkStart,
        end_offset: linkEnd,
      },
      version: {
        value: version.trim(),
        token_type: "version",
        start_offset: versionStart,
        end_offset: versionEnd,
      },
      path: {
        value: path.replace("path=", "").trim(),
        token_type: "path",
        start_offset: pathStart,
        end_offset: pathEnd,
      },
      user: {
        value: user.replace("user=", "").trim(),
        token_type: "user",
        start_offset: userStart,
        end_offset: userEnd,
      },
      token: {
        value: token.replace("token=", "").trim(),
        token_type: "token",
        start_offset: tokenStart,
        end_offset: tokenEnd,
      },
    };
  }
  return execResult;
}

export const UseItem = createToken({
  name: "UseItem",
  pattern: matchUseItem,
  line_breaks: false,
});

function matchVar(text: string) {
  const varPattern =
    /^[ ]*var([ ]+\S+)([ ]+(?:uses|network|var))?([ ]+\S+)\s*$/;
  const execResult = varPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const varName = execResult[1];
    let varReferenceType = "";
    if (execResult[2] !== undefined) {
      varReferenceType = execResult[2];
    }
    const varValue = execResult[3];
    const varNameStart = execResult.index + "var".length + 1;
    const varNameEnd = varNameStart + varName.length;
    let varReferenceTypeStart = null;
    let varReferenceTypeEnd = null;
    if (varReferenceType !== "") {
      varReferenceTypeStart =
        execResult.index + "var".length + varName.length + 1;
      varReferenceTypeEnd = varReferenceTypeStart + varReferenceType.length;
    }
    const varValueStart =
      execResult.index +
      "var".length +
      varName.length +
      varReferenceType.length +
      1;
    const varValueEnd = varValueStart + varValue.length;
    execResult.payload = {
      var_name: {
        value: varName.trim(),
        token_type: "variable_name",
        start_offset: varNameStart,
        end_offset: varNameEnd,
      },
      var_reference_type: {
        value: varReferenceType.trim(),
        token_type: "variable_reference_type",
        start_offset: varReferenceTypeStart,
        end_offset: varReferenceTypeEnd,
      },
      var_value: {
        value: varValue.trim(),
        token_type: "variable_value",
        start_offset: varValueStart,
        end_offset: varValueEnd,
      },
    };
  }
  return execResult;
}

export const Var = createToken({
  name: "Var",
  pattern: matchVar,
  line_breaks: false,
});

function matchModel(text: string) {
  // model model_name repo_name [arch=linux-x86] [uid=41]
  let modelPattern =
    /^model[ ]+(?!arch=|uid=|external=)(\S+)[ ]+(?!arch=|uid=|external=)(\S+)([ ]+arch=\S+)?([ ]+uid=\d+)?\s*$/;
  let execResult = modelPattern.exec(text) as CustomRegExpExecArray;

  let modelName = "";
  let modelRepoValue = "";
  let modelArch = "";
  let modelUid = "";
  let modelNameStart: number = 0;
  let modelNameEnd: number = 0;
  let modelRepoValueStart: number = 0;
  let modelRepoValueEnd: number = 0;
  let modelArchValueStart = null;
  let modelArchValueEnd = null;
  let modelUidValueStart = null;
  let modelUidValueEnd = null;
  let external = "false";
  let externalValueStart = null;
  let externalValueEnd = null;
  if (execResult == null) {
    if (text.includes("external=")) {
      // model model_name external=true [arch=linux-x86] [uid=41]
      modelPattern =
        /^model[ ]+(?!arch=|uid=|external=)(\S+)(?!arch=|uid=)([ ]+external=\w+)([ ]+arch=\S+)?([ ]+uid=\d+)?\s*$/;
      execResult = modelPattern.exec(text) as CustomRegExpExecArray;
      if (execResult != null) {
        modelName = execResult[1];
        modelRepoValue = "";
        external = execResult[2];
        modelArch = "";
        if (execResult[3] !== undefined) {
          modelArch = execResult[3];
        } else if (parsedStackArch !== "") {
          modelArch = parsedStackArch;
        } else {
          modelArch = defaultArch;
        }
        modelUid = "";
        if (execResult[4] !== undefined) {
          modelUid = execResult[4];
        }
        modelNameStart = execResult.index + "model".length + 1;
        modelNameEnd = modelNameStart + modelName.length;
        externalValueStart =
          execResult.index + "model".length + 1 + modelName.length + 1;
        externalValueEnd = externalValueStart + modelArch.length;
        modelArchValueStart = null;
        modelArchValueEnd = null;
        if (modelArch !== "") {
          modelArchValueStart =
            execResult.index +
            "model".length +
            1 +
            modelName.length +
            modelArch.length +
            1;
          modelArchValueEnd = modelArchValueStart + modelArch.length;
        }
        modelUidValueStart = null;
        modelUidValueEnd = null;
        if (modelUid !== "") {
          modelUidValueStart =
            execResult.index +
            "model".length +
            1 +
            modelName.length +
            modelArch.length +
            modelArch.length +
            1;
          modelUidValueEnd = modelUidValueStart + modelUid.length;
        }
      } else {
        // model model_name repo_name external=true [arch=linux-x86] [uid=41]
        modelPattern =
          /^model[ ]+(?!arch=|uid=|external=)(\S+)[ ]+(?!arch=|uid=)(\S+)([ ]+external=\w+)([ ]+arch=\S+)?([ ]+uid=\d+)?\s*$/;
        execResult = modelPattern.exec(text) as CustomRegExpExecArray;
        if (execResult != null) {
          modelName = execResult[1];
          modelRepoValue = execResult[2];
          external = execResult[3];
          modelArch = "";
          if (execResult[4] !== undefined) {
            modelArch = execResult[4];
          } else if (parsedStackArch !== "") {
            modelArch = parsedStackArch;
          } else {
            modelArch = defaultArch;
          }
          modelUid = "";
          if (execResult[5] !== undefined) {
            modelUid = execResult[5];
          }
          modelNameStart = execResult.index + "model".length + 1;
          modelNameEnd = modelNameStart + modelName.length;
          modelRepoValueStart =
            execResult.index + "model".length + 1 + modelName.length + 1;
          modelRepoValueEnd = modelRepoValueStart + modelRepoValue.length;
          externalValueStart =
            execResult.index +
            "model".length +
            1 +
            modelName.length +
            modelRepoValue.length +
            1;
          externalValueEnd = externalValueStart + external.length;
          modelArchValueStart = null;
          modelArchValueEnd = null;
          if (modelArch !== "") {
            modelArchValueStart =
              execResult.index +
              "model".length +
              1 +
              modelName.length +
              modelRepoValue.length +
              external.length +
              1;
            modelArchValueEnd = modelArchValueStart + modelArch.length;
          }
          modelUidValueStart = null;
          modelUidValueEnd = null;
          if (modelUid !== "") {
            modelUidValueStart =
              execResult.index +
              "model".length +
              1 +
              modelName.length +
              modelRepoValue.length +
              external.length +
              modelArch.length +
              1;
            modelUidValueEnd = modelUidValueStart + modelUid.length;
          }
        }
      }
    }
  } else {
    modelName = execResult[1];
    modelRepoValue = execResult[2];
    modelArch = "";
    if (execResult[3] !== undefined) {
      modelArch = execResult[3];
    } else if (parsedStackArch !== "") {
      modelArch = parsedStackArch;
    } else {
      modelArch = defaultArch;
    }
    modelUid = "";
    if (execResult[4] !== undefined) {
      modelUid = execResult[4];
    }
    modelNameStart = execResult.index + "model".length + 1;
    modelNameEnd = modelNameStart + modelName.length;
    modelRepoValueStart =
      execResult.index + "model".length + 1 + modelName.length + 1;
    modelRepoValueEnd = modelRepoValueStart + modelRepoValue.length;
    modelArchValueStart = null;
    modelArchValueEnd = null;
    if (modelArch !== "") {
      modelArchValueStart =
        execResult.index +
        "model".length +
        1 +
        modelName.length +
        modelRepoValue.length +
        1;
      modelArchValueEnd = modelArchValueStart + modelArch.length;
    }
    modelUidValueStart = null;
    modelUidValueEnd = null;
    if (modelUid !== "") {
      modelUidValueStart =
        execResult.index +
        "model".length +
        1 +
        modelName.length +
        modelRepoValue.length +
        modelArch.length +
        1;
      modelUidValueEnd = modelUidValueStart + modelUid.length;
    }
  }
  if (execResult != null) {
    execResult.payload = {
      model_name: {
        value: modelName.trim(),
        token_type: "model_name",
        start_offset: modelNameStart,
        end_offset: modelNameEnd,
      },
      model_repo_name: {
        value: modelRepoValue.trim(),
        token_type: "model_repo_name",
        start_offset: modelRepoValueStart,
        end_offset: modelRepoValueEnd,
      },
      model_arch: {
        value: modelArch.replace("arch=", "").trim(),
        token_type: "model_arch",
        start_offset: modelArchValueStart,
        end_offset: modelArchValueEnd,
      },
      model_uid: {
        value: modelUid.replace("uid=", "").trim(),
        token_type: "model_uid",
        start_offset: modelUidValueStart,
        end_offset: modelUidValueEnd,
      },
      external: {
        value: external.replace("external=", "").trim(),
        token_type: "external",
        start_offset: externalValueStart,
        end_offset: externalValueEnd,
      },
    };
  }
  return execResult;
}

export const Model = createToken({
  name: "Model",
  pattern: matchModel,
  line_breaks: false,
});

function matchEnvVar(text: string) {
  const varPattern = /^envar([ ]+\S+)([ ]+\S+)\s*$/;
  const execResult = varPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const varName = execResult[1];
    const varValue = execResult[2];
    const varNameStart = execResult.index + "envar".length + 1;
    const varNameEnd = varNameStart + varName.length;
    const varValueStart =
      execResult.index + "envar".length + varName.length + 1;
    const varValueEnd = varValueStart + varValue.length;
    execResult.payload = {
      env_var_name: {
        value: varName.trim(),
        token_type: "env_variable_name",
        start_offset: varNameStart,
        end_offset: varNameEnd,
      },
      env_var_value: {
        value: varValue.trim(),
        token_type: "env_variable_value",
        start_offset: varValueStart,
        end_offset: varValueEnd,
      },
    };
  }
  return execResult;
}

export const EnvVar = createToken({
  name: "EnvVar",
  pattern: matchEnvVar,
  line_breaks: false,
});

function matchWorkflow(text: string) {
  const workflowPattern = /^workflow([ ]+\S+)\s*$/;
  const execResult = workflowPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const workflowName = execResult[1];
    const workflowNameStart = execResult.index + "workflow".length + 1;
    const workflowNameEnd = workflowNameStart + workflowName.length;
    execResult.payload = {
      workflow_name: {
        value: workflowName.trim(),
        token_type: "workflow_name",
        start_offset: workflowNameStart,
        end_offset: workflowNameEnd,
      },
    };
  }
  return execResult;
}

export const Workflow = createToken({
  name: "Workflow",
  pattern: matchWorkflow,
  line_breaks: false,
});

function matchStack(text: string) {
  const stackPattern =
    /^stack[ ]+(?!stacked=|sequential=|arch=)(\S+)([ ]+stacked=(?:true|false))?([ ]+sequential=(?:true|false))?([ ]+arch=\S+)?\s*$/;
  const execResult = stackPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    parsedStackArch = "";
    const stackName = execResult[1];
    let stacked = "";
    if (execResult[2] !== undefined) {
      stacked = execResult[2];
    }
    let sequential = "";
    if (execResult[3] !== undefined) {
      sequential = execResult[3];
    }
    let stackArch = "";
    if (execResult[4] !== undefined) {
      stackArch = execResult[4];
      parsedStackArch = stackArch;
    } else {
      stackArch = defaultArch;
    }
    const stackNameStart = execResult.index + "stack".length + 1;
    const stackNameEnd = stackNameStart + stackName.length;
    const stackedStart =
      execResult.index + "stack".length + stackName.length + 1;
    const stackedEnd = stackedStart + stacked.length;
    const sequentialStart =
      execResult.index + "stack".length + stackName.length + stacked.length + 1;
    const sequentialEnd = sequentialStart + sequential.length;
    const stackArchStart =
      execResult.index +
      "stack".length +
      stackName.length +
      stacked.length +
      sequential.length +
      1;
    const stackArchEnd = stackArchStart + stackArch.length;
    execResult.payload = {
      stack_name: {
        value: stackName.trim(),
        token_type: "stack_name",
        start_offset: stackNameStart,
        end_offset: stackNameEnd,
      },
      stacked: {
        value: stacked.replace("stacked=", "").trim(),
        token_type: "stacked",
        start_offset: stackedStart,
        end_offset: stackedEnd,
      },
      sequential: {
        value: sequential.replace("sequential=", "").trim(),
        token_type: "sequential",
        start_offset: sequentialStart,
        end_offset: sequentialEnd,
      },
      stack_arch: {
        value: stackArch.replace("arch=", "").trim(),
        token_type: "stack_arch",
        start_offset: stackArchStart,
        end_offset: stackArchEnd,
      },
    };
  }
  return execResult;
}

export const Stack = createToken({
  name: "Stack",
  pattern: matchStack,
  line_breaks: false,
});

export const WhiteSpace = createToken({
  name: "WhiteSpace",
  pattern: /[ \t\n\r]+/,
  group: Lexer.SKIPPED, // Skip whitespace.
});

// Define the lexer with tokens in the proper order.
export const allTokens = [
  WhiteSpace,
  Simulation,
  Channel,
  Network,
  File,
  Uses,
  UseItem,
  Var,
  Stack,
  Model,
  EnvVar,
  Workflow,
];

export const fsilLexer = new Lexer(allTokens);
export function lex(inputText: string) {
  const lines = inputText.split(/\r?\n/);
  let lineStartOffset = 0;
  let lexingResult = {
    tokens: [] as IToken[],
    groups: {} as Record<string, IToken[]>,
    errors: [] as ILexingError[],
  };
  lines.forEach((line) => {
    if (line !== "") {
      const res = fsilLexer.tokenize(line);
      lexingResult["tokens"].push(...res["tokens"]);
      Object.assign(lexingResult["groups"], res["groups"]);
      lexingResult["errors"].push(...res["errors"]);
      if (lexingResult.errors.length > 0) {
        console.log(lexingResult.errors);
        throw Error("Lexing errors detected");
      }
      lineStartOffset += line.length + 1;
    }
  });
  return lexingResult;
}
