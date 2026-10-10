echo "Starting compilation..."

if [ -f "./build.sh" ]; then
    echo "📦 Triggering build script..."
    BUILT_FILE=$(./build.sh "1.0.0")
else
    echo "build.sh not found!!!"
    exit 1
fi

echo "Starting Docker daemon"
sudo systemctl start docker

chmod +x "$BUILT_FILE"
./"$BUILT_FILE"

echo "Deployment completed!"

