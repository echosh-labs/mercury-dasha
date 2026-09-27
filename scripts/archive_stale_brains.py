import os
import shutil

# 1. Archive ~/.gemini/antigravity/brain
wsl_brain = os.path.expanduser('~/.gemini/antigravity/brain')
wsl_archive = os.path.expanduser('~/.gemini/antigravity/archive_pre_august')
os.makedirs(wsl_archive, exist_ok=True)

if os.path.exists(wsl_brain):
    for entry in os.listdir(wsl_brain):
        src_dir = os.path.join(wsl_brain, entry)
        if os.path.isdir(src_dir):
            dst_dir = os.path.join(wsl_archive, entry)
            if not os.path.exists(dst_dir):
                shutil.move(src_dir, dst_dir)
                print(f'Archived WSL brain: {entry}')

# 2. Archive pre-August Windows brain sessions in /mnt/c/Users/justi/.gemini/antigravity/brain
win_brain = '/mnt/c/Users/justi/.gemini/antigravity/brain'
win_archive = '/mnt/c/Users/justi/.gemini/antigravity/archive_pre_august'
os.makedirs(win_archive, exist_ok=True)

pre_august = [
    '264ddd00-f86f-45a9-bad7-a7265f57cfb0',
    '382d4fec-ba2c-4b41-8888-d80f2bd67bd7',
    '5beb3dd2-d0dd-4f7e-92b2-990d9b96dcab',
    '6b220c01-1d79-4441-93c3-4eb7349ff0f5',
    '7c12f7ce-3a12-4d1d-960e-64c1b92975e4',
    '893920ff-8498-4ad0-b73a-06c2ddb6fbce',
    '9017128d-cf2e-499c-a2f1-434eb4e86fd2',
    'be5d7031-4e03-4a5a-b122-274851814532',
    'd2dacd0c-96b5-4d59-ab85-d166b79d1e00',
    'fe360956-b3ee-4254-b267-04e1493bf795'
]

if os.path.exists(win_brain):
    for entry in pre_august:
        src_dir = os.path.join(win_brain, entry)
        if os.path.isdir(src_dir):
            dst_dir = os.path.join(win_archive, entry)
            if not os.path.exists(dst_dir):
                shutil.move(src_dir, dst_dir)
                print(f'Archived Windows brain: {entry}')

print('Archive completed successfully.')
