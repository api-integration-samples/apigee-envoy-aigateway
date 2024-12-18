
var geminiResponse = JSON.parse(response.content);
var questionResponse = convertResponse(geminiResponse);

if (geminiResponse.length > 0 && geminiResponse[geminiResponse.length - 1].usageMetadata) {
  var usageMetadata = geminiResponse[geminiResponse.length - 1].usageMetadata;
  context.setVariable("genai.promptTokenCount", usageMetadata.promptTokenCount);
  context.setVariable("genai.responseTokenCount", usageMetadata.candidatesTokenCount);
  context.setVariable("genai.totalTokenCount", usageMetadata.totalTokenCount);
  context.setVariable("genai.modelName", "/gemini15-pro");
}

context.setVariable("response.content", JSON.stringify({
  "prompt": context.getVariable("genai.prompt"),
  "answer": questionResponse,
  "usage": {
    "prompt_tokens": context.getVariable("genai.promptTokenCount"),
    "completion_tokens": context.getVariable("genai.responseTokenCount"),
    "total_tokens": context.getVariable("genai.totalTokenCount")
  }
}));

function convertResponse(dataResponseObject) {
  var result = "";

  for (i = 0; i < dataResponseObject.length; i++) {
    result += dataResponseObject[i]["candidates"][0]["content"]["parts"][0]["text"];
  }

  return result;
}

// this is to only export the function if in node
if (typeof exports !== 'undefined') {
  exports.convertResponse = convertResponse;
}