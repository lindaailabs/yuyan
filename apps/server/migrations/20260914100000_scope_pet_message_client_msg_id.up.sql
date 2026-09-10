ALTER TABLE pet_messages
  DROP INDEX uk_client_msg_id,
  ADD UNIQUE KEY uk_user_client_msg_id (user_id, client_msg_id);
