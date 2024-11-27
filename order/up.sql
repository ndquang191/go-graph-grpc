CREATE TABLE IF NOT EXISTS `order` (
   id char(27) PRIMARY KEY,
   created_at TIMESTAMP WITH TIME ZONE NOT NULL,
   account_id char(27) NOT NULL,
   total_price MONEY NOT NULL,
);

CREATE TABLE IF NOT EXISTS `order_products` (
   order_id char(27) REFERENCES `order`(id) ON DELETE CASCADE,
   product_id char(27),
   quantity INT NOT NULL,
   PRIMARY KEY (order_id, product_id)
);

