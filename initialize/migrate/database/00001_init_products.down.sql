-- 00001_init_products.down.sql
DELETE FROM products WHERE slug IN (
    'gemini-advanced-18m',
    'chatgpt-plus-personal',
    'claude-pro-ready',
    'cursor-pro-1m'
);
