INSERT INTO pmsn.resource (id, code, name, description) VALUES
('res-1', 'order', 'Order', 'Order resource'),
('res-2', 'product', 'Product', 'Product resource')
ON CONFLICT DO NOTHING;

INSERT INTO pmsn.action (id, resource_id, code, name, description) VALUES
('act-1', 'res-1', 'read', 'Read Order', 'Read order details'),
('act-2', 'res-1', 'write', 'Write Order', 'Create or update order'),
('act-3', 'res-2', 'read', 'Read Product', 'Read product details'),
('act-4', 'res-2', 'write', 'Write Product', 'Create or update product')
ON CONFLICT DO NOTHING;
