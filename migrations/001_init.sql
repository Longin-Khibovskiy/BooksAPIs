CREATE TABLE IF NOT EXISTS books (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    author TEXT NOT NULL,
    description TEXT,
    publisher TEXT,
    image TEXT,
    amazon_url TEXT,
    rank INT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS book_links (
    id SERIAL PRIMARY KEY,
    book_id INT REFERENCES books(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    url TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_books_rank ON books(rank);
CREATE INDEX IF NOT EXISTS idx_book_links_book_id ON book_links(book_id);
