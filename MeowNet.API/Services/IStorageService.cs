namespace MeowNet.API.Services
{
    public interface IStorageService
    {
        Task SaveFileAsync(string fileName, Stream stream, string subFolder = "");
        Task<byte[]> GetFileAsync(string fileName, string subFolder = "");
        Task DeleteFileAsync(string fileName, string subFolder = "");
    }
}
