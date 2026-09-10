ALTER TABLE pet_messages
  DROP INDEX uk_user_client_msg_id,
  ADD UNIQUE KEY uk_client_msg_id (client_msg_id);
