@echo off
echo Fixing git configuration...

echo 2. Removing potential SkillSync submodule tracking...
git rm --cached SkillSync

echo 3. Removing nested .git directory in SkillSync/SkillSync...
rmdir /s /q "SkillSync\SkillSync\.git"
if exist "SkillSync\.git" rmdir /s /q "SkillSync\.git"

echo.
echo ==========================================
echo Fix applied. Now please run these commands:
echo.
echo 1. git add .
echo 2. git commit -m "Fix project structure and remove secrets"
echo 3. git push
echo ==========================================
pause
