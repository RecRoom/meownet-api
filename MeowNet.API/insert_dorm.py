import sqlite3
from datetime import datetime

conn = sqlite3.connect('meownet.db')
c = conn.cursor()

c.execute("SELECT room_id FROM Rooms WHERE name='DormRoom'")
if not c.fetchone():
    print("Inserting DormRoom...")
    c.execute("""
        INSERT INTO Rooms (
            name, description, creator_account_id, is_dorm, supports_vr_low, supports_mobile, 
            supports_screens, supports_walk_vr, supports_teleport_vr, supports_juniors, 
            max_players, persistence_version, ugc_version, created_at,
            image_name, custom_warning, ranked_entity_id, state, accessibility
        ) VALUES (
            'DormRoom', 'Your personal room', 1, 1, 1, 1, 1, 1, 1, 1, 4, 1, 1, ?, '', '', '', 0, 1
        )
    """, (datetime.now().isoformat(),))
    
    room_id = c.lastrowid
    print(f"Inserted Room with ID {room_id}")
    
    c.execute("""
        INSERT INTO SubRooms (
            room_id, sub_room_id, name, is_sandbox, max_players, accessibility, saved_by_account_id, unity_scene_id, data_blob
        ) VALUES (
            ?, 1, 'DormRoom', 0, 4, 0, 1, '', ''
        )
    """, (room_id,))
    conn.commit()
    print("Done!")
conn.close()
