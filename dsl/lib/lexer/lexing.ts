// Copyright 2024 Robert Bosch GmbH
//
// SPDX-License-Identifier: Apache-2.0

import { Lexer, createToken, IToken, ILexingError } from "chevrotain";

let defaultArch = "linux-amd64";
const defaultStepsize = "0.0005";
const defaultEndtime = "0.005";
let parsedStackArch = "";
interface CustomRegExpExecArray extends RegExpExecArray {
  payload?: any;
}


function matchSimulation(text: string) {
  const simulationPattern =
    /^[ \t]*simulation([ ]+arch\=\S+)?([ ]+stepsize\=(?:\d*\.?\d+))?([ ]+endtime\=(?:\d*\.?\d+))?\s*(?:\#.*)?$/;
  const execResult = simulationPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    let simulationArch = defaultArch;
    if (execResult[1] !== undefined) {
      simulationArch = execResult[1];
      defaultArch = simulationArch.replace("arch=", "").trim();
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

    execResult.payload = {
      simulation_arch: {
        value: defaultArch,
        token_type: "simulation_arch",
      },
      stepsize: {
        value: stepsize.replace("stepsize=", "").trim(),
        token_type: "stepsize",
      },
      endtime: {
        value: endtime.replace("endtime=", "").trim(),
        token_type: "endtime",
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
  const channelPattern = /^[ \t]*channel([ ]+\w+)(?:([ ]+\w+))?\s*(?:\#.*)?$/;
  const execResult = channelPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const channelName = execResult[1];
    let channelAlias = "";
    if (execResult[2] !== undefined) {
      channelAlias = execResult[2];
    }

    execResult.payload = {
      channel_name: {
        value: channelName.trim(),
        token_type: "channel_name",
      },
      channel_alias: {
        value: channelAlias.trim(),
        token_type: "channel_alias",
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
  const filePattern = /^[ \t]*file([ ]+\S+)([ ]+(?:uses))?([ ]+\S+)?\s*(?:\#.*)?$/;
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

    execResult.payload = {
      file_name: {
        value: fileName.trim(),
        token_type: "file_name",
      },
      file_reference_type: {
        value: fileReferenceType.trim(),
        token_type: "file_reference_type",
      },
      file_value: {
        value: fileValue.trim(),
        token_type: "file_value",
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
  const networkPattern = /^[ \t]*network([ ]+\S+)([ ]+\'\S+\')\s*(?:\#.*)?$/;
  const execResult = networkPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const networkName = execResult[1];
    const mimeType = execResult[2];
    execResult.payload = {
      network_name: {
        value: networkName.trim(),
        token_type: "network_name",
      },
      mime_type: {
        value: mimeType.replace(/\'/g, "").trim(),
        token_type: "mime_type",
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
  const usesKeywordPattern = /^[ \t]*uses[ \t]*(?:\#.*)?$/;
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
    /^[ \t]*(\S+)([ ]+(?:(?:https\:\/\/\S+)|(?:\S+\.\S+)|(?:\S+\/\S+)))([ ]+v\d+(?:\.\d+)*)?(?:[ ]+(path\=\S+))?(?:[ ]+(user\=\S+))?(?:[ ]+(token\=\S+))?\s*(?:\#.*)?$/;
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

    execResult.payload = {
      use_item: {
        value: useItem.trim(),
        token_type: "use_item",
      },
      link: {
        value: link.trim(),
        token_type: "link",
      },
      version: {
        value: version.trim(),
        token_type: "version",
      },
      path: {
        value: path.replace("path=", "").trim(),
        token_type: "path",
      },
      user: {
        value: user.replace("user=", "").trim(),
        token_type: "user",
      },
      token: {
        value: token.replace("token=", "").trim(),
        token_type: "token",
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
    /^[ \t]*var([ ]+\S+)([ ]+(?:uses|network|var))?([ ]+(?:\S+([ ]+(?:mimetype|signal))?|(?:\".*\")|(?:\'.*\')))\s*(?:\#.*)?$/;
  const execResult = varPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const varName = execResult[1];
    let varReferenceType = "";
    if (execResult[2] !== undefined) {
      varReferenceType = execResult[2];
    }
    let varValue = execResult[3].trim();
    let varNetworkType = "";
    if (execResult[4] !== undefined) {
      varNetworkType = execResult[4].trim();
      if (varNetworkType != "") {
        if (varValue.endsWith("mimetype")) {
          varValue = varValue.replace(/mimetype$/, "").trim();
        } else if (varValue.endsWith("signal")) {
          varValue = varValue.replace(/signal$/, "").trim();
        }
      }
    }

    execResult.payload = {
      var_name: {
        value: varName.trim(),
        token_type: "variable_name",
      },
      var_reference_type: {
        value: varReferenceType.trim(),
        token_type: "variable_reference_type",
      },
      var_value: {
        value: varValue,
        token_type: "variable_value",
      },
      var_network_type: {
        value: varNetworkType,
        token_type: "variable_network_type",
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
    /^[ \t]*model[ ]+(?!arch=|uid=|external=)(\S+)[ ]+(?!arch=|uid=|external=)(\S+)([ ]+arch=\S+)?([ ]+uid=\d+)?\s*(?:\#.*)?$/;
  let execResult = modelPattern.exec(text) as CustomRegExpExecArray;

  let modelName = "";
  let modelRepoValue = "";
  let modelArch = "";
  let modelUid = "";
  let external = "false";
  if (execResult == null) {
    if (text.includes("external=")) {
      // model model_name external=true [arch=linux-x86] [uid=41]
      modelPattern =
        /^[ \t]*model[ ]+(?!arch=|uid=|external=)(\S+)(?!arch=|uid=)([ ]+external=\w+)([ ]+arch=\S+)?([ ]+uid=\d+)?\s*(?:\#.*)?$/;
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
      } else {
        // model model_name repo_name external=true [arch=linux-x86] [uid=41]
        modelPattern =
          /^[ \t]*model[ ]+(?!arch=|uid=|external=)(\S+)[ ]+(?!arch=|uid=)(\S+)([ ]+external=\w+)([ ]+arch=\S+)?([ ]+uid=\d+)?\s*(?:\#.*)?$/;
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
  }
  if (execResult != null) {
    execResult.payload = {
      model_name: {
        value: modelName.trim(),
        token_type: "model_name",
      },
      model_repo_name: {
        value: modelRepoValue.trim(),
        token_type: "model_repo_name",
      },
      model_arch: {
        value: modelArch.replace("arch=", "").trim(),
        token_type: "model_arch",
      },
      model_uid: {
        value: modelUid.replace("uid=", "").trim(),
        token_type: "model_uid",
      },
      external: {
        value: external.replace("external=", "").trim(),
        token_type: "external",
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
  const varPattern = /^[ \t]*envar([ ]+\S+)([ ]+\S+)\s*(?:\#.*)?$/;
  const execResult = varPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const varName = execResult[1];
    const varValue = execResult[2];
    execResult.payload = {
      env_var_name: {
        value: varName.trim(),
        token_type: "env_variable_name",
      },
      env_var_value: {
        value: varValue.trim(),
        token_type: "env_variable_value",
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

function matchAnnotation(text: string) {
  const annotationPattern = /^[ \t]*annotation([ ]+\S+)([ ]+\S+)\s*(?:\#.*)?$/;
  const execResult = annotationPattern.exec(text) as CustomRegExpExecArray;
  if (execResult !== null) {
    const annotationName = execResult[1];
    const annotationValue = execResult[2];
    execResult.payload = {
      annotation_name: {
        value: annotationName.trim(),
        token_type: "annotation_name",
      },
      annotation_value: {
        value: annotationValue.trim(),
        token_type: "annotation_value",
      },
    };
  }
  return execResult;
}

export const Annotation = createToken({
  name: "Annotation",
  pattern: matchAnnotation,
  line_breaks: false,
});

function matchWorkflow(text: string) {
  const workflowPattern = /^[ \t]*workflow([ ]+\S+)(?:([ ]+uses)([ ]+\S+))?\s*(?:\#.*)?$/;
  const execResult = workflowPattern.exec(text) as CustomRegExpExecArray;
  let workflowReferenceType = "";
  let workflowValue = "";
  if (execResult !== null) {
    const workflowName = execResult[1];

    workflowReferenceType = "";
    if (execResult[2] !== undefined) {
      workflowReferenceType = execResult[2];
    }

    workflowValue = "";
    if (execResult[3] !== undefined) {
      workflowValue = execResult[3];
    }
    execResult.payload = {
      workflow_name: {
        value: workflowName.trim(),
        token_type: "workflow_name",
      },
      workflow_reference_type: {
        value: workflowReferenceType.trim(),
        token_type: "workflow_reference_type",
      },
      workflow_value: {
        value: workflowValue.trim(),
        token_type: "workflow_value",
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
    /^[ \t]*stack[ ]+(?!stacked=|sequential=|arch=)(\S+)([ ]+stacked=(?:true|false))?([ ]+sequential=(?:true|false))?([ ]+arch=\S+)?\s*(?:\#.*)?$/;
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

    execResult.payload = {
      stack_name: {
        value: stackName.trim(),
        token_type: "stack_name",
      },
      stacked: {
        value: stacked.replace("stacked=", "").trim(),
        token_type: "stacked",
      },
      sequential: {
        value: sequential.replace("sequential=", "").trim(),
        token_type: "sequential",
      },
      stack_arch: {
        value: stackArch.replace("arch=", "").trim(),
        token_type: "stack_arch",
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

const Comment = createToken({
  name: "Comment",
  pattern: /#.*/,
  group: Lexer.SKIPPED,
});

// Define the lexer with tokens in the proper order.
export const allTokens = [
  WhiteSpace,
  Comment,
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
  Annotation,
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
