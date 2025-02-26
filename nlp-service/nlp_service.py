import os
import spacy
import torch
from transformers import pipeline, AutoTokenizer, AutoModelForSeq2SeqLM
from flask import Flask, request, jsonify
from sentence_transformers import SentenceTransformer

app = Flask(__name__)

# Load spaCy (transformer-based) - may change for deployment
spacy_model = spacy.load("en_core_web_trf")  # or "en_core_web_sm"

# Huggingface summarizer
summarizer = pipeline(
    "summarization",
    model="facebook/bart-large-cnn",
    tokenizer="facebook/bart-large-cnn",
    framework="pt",  # "tf" if using TensorFlow
    device=0 if torch.cuda.is_available() else -1
)

# Sentence Transformer for embeddings
embedding_model = SentenceTransformer("all-MiniLM-L6-v2")

@app.route("/nlp", methods=["POST"])
def nlp_process():
    data = request.get_json()
    text = data.get("text", "")
    if not text.strip():
        return jsonify({
            "entities": [],
            "keywords": [],
            "summary": "",
            "embedding": []
        })

    doc = spacy_model(text)

    # Named entities (spacy-based)
    entities = []
    for ent in doc.ents:
        entities.append({
            "text": ent.text,
            "label": ent.label_
        })

    # Simple keyword extraction with noun chunks
    # For something more advanced: PyTextRank, RAKE, or custom pipeline
    keywords = list(set(chunk.text for chunk in doc.noun_chunks if chunk.text.strip()))

    # Summarization with huggingface
    # We might chunk the text if it's very long
    MAX_LEN = 1024  # bart-large-cnn often limited to 1024 tokens
    # If text is too long, we truncate or chunk
    if len(text.split()) > 1500:
        # chunk or reduce
        text = " ".join(text.split()[:1500])

    summary_result = summarizer(text, max_length=100, min_length=30, do_sample=False)
    summary_text = summary_result[0]["summary_text"] if summary_result else ""


    # Embeddings
    embedding = embedding_model.encode(text, show_progress_bar=False).tolist()

    return jsonify({
        "entities": entities,
        "keywords": keywords,
        "summary": summary_text,
        "embedding": embedding
    })

if __name__ == "__main__":
    app.run(host="0.0.0.0", port=5000)
