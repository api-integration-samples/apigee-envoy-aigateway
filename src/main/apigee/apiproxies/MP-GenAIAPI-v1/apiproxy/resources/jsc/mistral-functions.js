var mistralResponse = JSON.parse(response.content);

mistralResponse["prompt"] = context.getVariable("genai.prompt");

context.setVariable("response.content", JSON.stringify(mistralResponse));