export interface Product {
  id: string;
  name: string;
  description: string;
  category?: string;
  price_cents: number;
  sku: string;
  stock: number;
  thumbnail?: string;
  images?: string;
  rating?: number;
  review_count?: number;
  created_at: string;
  updated_at: string;
}

export interface CartItem {
  product_id: string;
  quantity: number;
}

export interface Cart {
  user_id: string;
  items: CartItem[];
}

export interface Order {
  id: string;
  order_id?: string;
  user_id: string;
  amount_cents: number;
  status: string;
  payment_ref: string;
  created_at: string;
  items: CartItem[];
}
