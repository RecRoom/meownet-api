import sqlite3
import json
import datetime
conn = sqlite3.connect('meownet.db')
cursor = conn.cursor()
with open('../db/seeds/defaultrooms.json', 'r', encoding='utf-8') as f:
    rooms = json.load(f)
for r in rooms:
    cursor.execute('SELECT 1 FROM Rooms WHERE room_id = ?', (r['RoomId'],))
    if cursor.fetchone(): continue
    cursor.execute('INSERT INTO Rooms (room_id, name, description, image_name, creator_account_id, state, accessibility, auto_localize_room, cloning_allowed, custom_warning, disable_mic_auto_mute, disable_room_comments, encrypt_voice_chat, is_developer_owned, is_dorm, is_rro, load_screen_locked, max_player_calculation_mode, max_players, min_level, persistence_version, ranked_entity_id, ranking_context, supports_juniors, supports_level_voting, supports_mobile, supports_quest_2, supports_screens, supports_teleport_vr, supports_vr_low, supports_walk_vr, toxmod_enabled, ugc_version, warning_mask, data_blob, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)', (r.get('RoomId'), r.get('Name'), r.get('Description'), r.get('ImageName'), r.get('CreatorAccountId', 1), r.get('State', 0), r.get('Accessibility', 1), 1 if r.get('AutoLocalizeRoom') else 0, 1 if r.get('CloningAllowed') else 0, r.get('CustomWarning', ''), 1 if r.get('DisableMicAutoMute') else 0, 1 if r.get('DisableRoomComments') else 0, 1 if r.get('EncryptVoiceChat') else 0, 1 if r.get('IsDeveloperOwned') else 0, 1 if r.get('IsDorm') else 0, 1 if r.get('IsRRO') else 0, 1 if r.get('LoadScreenLocked') else 0, r.get('MaxPlayerCalculationMode', 1), r.get('MaxPlayers', 12), r.get('MinLevel', 0), r.get('PersistenceVersion', 2), r.get('RankedEntityId', ''), r.get('RankingContext', 0), 1 if r.get('SupportsJuniors') else 0, 1 if r.get('SupportsLevelVoting') else 0, 1 if r.get('SupportsMobile') else 0, 1 if r.get('SupportsQuest2') else 0, 1 if r.get('SupportsScreens') else 0, 1 if r.get('SupportsTeleportVR') else 0, 1 if r.get('SupportsVRLow') else 0, 1 if r.get('SupportsWalkVR') else 0, 1 if r.get('ToxmodEnabled') else 0, r.get('UgcVersion', 0), r.get('WarningMask', 0), r.get('DataBlob', ''), r.get('CreatedAt', datetime.datetime.utcnow().isoformat())))
    for sr in r.get('SubRooms', []):
        cursor.execute('SELECT 1 FROM SubRooms WHERE sub_room_id = ?', (sr['SubRoomId'],))
        if cursor.fetchone(): continue
        cursor.execute('INSERT INTO SubRooms (sub_room_id, room_id, accessibility, data_blob, is_sandbox, max_players, name, saved_by_account_id, unity_scene_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)', (sr.get('SubRoomId'), sr.get('RoomId', r['RoomId']), sr.get('Accessibility', 1), sr.get('DataBlob') or '', 1 if sr.get('IsSandbox') else 0, sr.get('MaxPlayers', 12), sr.get('Name'), sr.get('SavedByAccountId', 1), sr.get('UnitySceneId')))
conn.commit()
conn.close()
print('Rooms seeded successfully!')
