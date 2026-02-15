@echo off
echo Fixing git configuration...

echo 1. Removing main.py from git tracking (keeping local file)...
git rm --cached main.py

echo 2. Removing nested .git directory in SkillSync/SkillSync...
rmdir /s /q "SkillSync\SkillSync\.git"

echo.
echo ==========================================
echo Fix applied. Now please run these commands:
echo.
echo git add .
echo git commit -m "Fix project structure and remove secrets"
echo git push
echo ==========================================
pause
