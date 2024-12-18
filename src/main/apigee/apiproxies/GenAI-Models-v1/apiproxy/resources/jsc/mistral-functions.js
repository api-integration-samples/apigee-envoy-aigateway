var mistralResponse = JSON.parse(response.content);

mistralResponse["prompt"] = context.getVariable("genai.prompt");

context.setVariable("response.content", JSON.stringify(mistralResponse));
context.setVariable("genai.modelName", "/mistral2407-pixtral");