namespace MeowNet.API.Services
{
    public class LocalStorageService : IStorageService
    {
        private readonly string _basePath;

        public LocalStorageService(IWebHostEnvironment env)
        {
            _basePath = Path.Combine(env.ContentRootPath, "uploads");
            if (!Directory.Exists(_basePath))
            {
                Directory.CreateDirectory(_basePath);
            }
        }

        public async Task SaveFileAsync(string fileName, Stream stream, string subFolder = "")
        {
            var targetDir = Path.Combine(_basePath, subFolder);
            if (!Directory.Exists(targetDir))
            {
                Directory.CreateDirectory(targetDir);
            }

            var filePath = Path.Combine(targetDir, fileName);
            using var fileStream = new FileStream(filePath, FileMode.Create, FileAccess.Write);
            await stream.CopyToAsync(fileStream);
        }

        public async Task<byte[]> GetFileAsync(string fileName, string subFolder = "")
        {
            var filePath = Path.Combine(_basePath, subFolder, fileName);
            if (!File.Exists(filePath)) return Array.Empty<byte>();

            return await File.ReadAllBytesAsync(filePath);
        }

        public Task DeleteFileAsync(string fileName, string subFolder = "")
        {
            var filePath = Path.Combine(_basePath, subFolder, fileName);
            if (File.Exists(filePath))
            {
                File.Delete(filePath);
            }
            return Task.CompletedTask;
        }
    }
}
