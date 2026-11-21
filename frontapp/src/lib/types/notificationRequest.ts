export interface NotificationRequestMessage {
  id: number;
  request_id: number;
  sender_id: string;
  message: string;
  created_at: string;
  is_system: boolean;
  sender?: App.Locals['user'];
}

export interface NotificationRequest {
  id: number;
  title: string;
  body: string;
  target_text: string;
  status: 'pending' | 'approved' | 'rejected';
  requester_id: string;
  resolved_by?: string | null;
  resolved_at?: string | null;
  created_at: string;
  updated_at: string;
  requester?: App.Locals['user'];
  resolver?: App.Locals['user'];
  messages?: NotificationRequestMessage[];
}
