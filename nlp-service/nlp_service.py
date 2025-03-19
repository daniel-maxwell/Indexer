# nlp-service/nlp_service.py
import os
import spacy
import torch
import threading
from concurrent.futures import ThreadPoolExecutor
from flask import Flask, request, jsonify
from keybert import KeyBERT
from transformers import pipeline

app = Flask(__name__)

# Load spaCy model
model_name = os.environ.get('SPACY_MODEL', 'en_core_web_sm')
spacy_model = spacy.load(model_name)

# KeyBERT for keyword extraction
kw_model = KeyBERT()

# Configuration
MAX_BATCH_SIZE = int(os.environ.get('MAX_BATCH_SIZE', 20))
MAX_TEXT_LENGTH = int(os.environ.get('MAX_TEXT_LENGTH', 1500))

@app.route("/nlp", methods=["POST"])
def nlp_process_single():
    """Legacy endpoint for single document processing"""
    data = request.get_json()
    text = data.get("text", "")
    needs_summary = data.get("needs_summary", True)  # Default to True for backward compatibility
    
    if not text.strip():
        return jsonify({
            "entities": [],
            "keyphrases": []
        })
    
    result = process_document(text)
    return jsonify(result)

@app.route("/nlp/batch", methods=["POST"])
def nlp_process_batch():
    """New endpoint for batch processing multiple documents"""
    data = request.get_json()
    documents = data.get("documents", [])
    
    if not documents:
        return jsonify({"results": []})
    
    # Limit batch size to prevent overload
    batch_size = min(len(documents), MAX_BATCH_SIZE)
    documents = documents[:batch_size]
    
    # Extract texts and summary flags
    texts = [doc.get("text", "") for doc in documents]
    
    results = []
    
    # Process entities with spaCy's efficient pipe
    docs = list(spacy_model.pipe(texts))
    
    # Process all keyword extraction in parallel
    all_keywords = kw_model.extract_keywords(
        texts, 
        keyphrase_ngram_range=(1, 3), 
        use_mmr=True, 
        top_n=10
    )
    
    # Combine all results
    for i, doc in enumerate(docs):
        entities = [{"text": ent.text, "label": ent.label_} for ent in doc.ents]
        keyphrases = [k[0] for k in all_keywords[i]] if i < len(all_keywords) else []
        
        results.append({
            "entities": entities,
            "keyphrases": keyphrases,
        })
    
    return jsonify({"results": results})

def process_document(text):
    """Process a single document"""
    if not text.strip():
        return {
            "entities": [],
            "keyphrases": [],
        }
    
    # Named entity recognition
    doc = spacy_model(text)
    entities = [{"text": ent.text, "label": ent.label_} for ent in doc.ents]
    
    # Keyword extraction
    keywords = [k[0] for k in kw_model.extract_keywords(
        text, 
        keyphrase_ngram_range=(1, 3), 
        use_mmr=True, 
        top_n=10
    )]
    
    return {
        "entities": entities,
        "keyphrases": keywords
    }

@app.route("/health", methods=["GET"])
def health_check():
    """Health check endpoint"""
    return jsonify({"status": "ok"})

if __name__ == "__main__":
    host = os.environ.get("HOST", "0.0.0.0")
    port = int(os.environ.get("PORT", 5000))
    app.run(host=host, port=port)