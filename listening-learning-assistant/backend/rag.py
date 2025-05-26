import chromadb
# setup Chroma in-memory, for easy prototyping. Can add persistence easily!
client = chromadb.Client()

# Create collection. get_collection, get_or_create_collection, delete_collection also available!
collection = client.create_collection("jlptn5-listening-comprehension")

with open("./transcripts/doc1.txt", "r") as f, open("./transcripts/doc2.txt", "r") as g:
    doc1 = f.read()
    doc2 = g.read()
    
collection.add(
    documents=[doc1, doc2], # we handle tokenization, embedding, and indexing automatically. You can skip that and add your own embeddings as well
    metadatas=[
        {"source": "transcripts/doc1.txt"}, 
        {"source": "transcripts/doc2.txt"}
        ], # filter on these!
    ids=["doc1", "doc2"], # unique for each doc
)

# Query/search 2 most similar results. You can also .get by id
results = collection.query(
    query_texts=["This is a query document"],
    n_results=2,
    # where={"metadata_field": "is_equal_to_this"}, # optional filter
    # where_document={"$contains":"search_string"}  # optional filter
)
