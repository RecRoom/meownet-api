import sqlite3

conn = sqlite3.connect('meownet.db')
conn.execute("INSERT INTO SubRooms (room_id, name, is_sandbox, max_players, accessibility, saved_by_account_id, unity_scene_id, data_blob) VALUES (45, 'DormRoom', 0, 4, 0, 1, '', '')")
conn.commit()
conn.close()
print("Done!")
