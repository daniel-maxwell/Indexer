#!/usr/bin/env python3
import spacy
from flask import Flask, request, jsonify

app = Flask(__name__)

# Load spaCy model
nlp = spacy.load("en_core_web_sm")

@app.route("/nlp", methods=["POST"])
def nlp_process():
    data = request.get_json()
    text = data.get("text", "")
    doc = nlp(text)

    # Extract named entities
    entities = []
    for ent in doc.ents:
        entities.append({
            "text": ent.text,
            "label": ent.label_
        })

    # Will flesh this out further in the future. For now, pick nouns as a naive approach.
    keywords = [token.text for token in doc if token.pos_ == "NOUN"]

    # Will flesh this out further in the future. For now, just return the first sentence.
    summary = ""
    if len(list(doc.sents)) > 0:
        summary = list(doc.sents)[0].text

    return jsonify({
        "entities": entities,
        "keywords": keywords,
        "summary": summary
    })

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
