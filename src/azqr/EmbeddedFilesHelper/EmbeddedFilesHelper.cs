namespace azqr;

public static class EmbeddedFilesHelper
{
    public static Dictionary<string, string> GetTemplates(string folder)
    {
        var results = new Dictionary<string, string>();
        var embeddedProvider = new EmbeddedFileProvider(Assembly.GetExecutingAssembly());
        var fileInfo = embeddedProvider.GetDirectoryContents(string.Empty).Where(f => f.Name.StartsWith(folder));

        foreach (var file in fileInfo)
        {
            using (var stream = file.CreateReadStream())
            {
                using (var reader = new StreamReader(stream))
                {
                    results.Add(file.Name, reader.ReadToEnd());
                }
            }
        }

        return results;
    }

    public static string GetTemplate(string templateName)
    {
        var embeddedProvider = new EmbeddedFileProvider(Assembly.GetExecutingAssembly());
        var fileInfo = embeddedProvider.GetFileInfo(templateName);
        if (fileInfo == null || !fileInfo.Exists)
            return string.Empty;

        using (var stream = fileInfo.CreateReadStream())
        {
            using (var reader = new StreamReader(stream))
            {
                return reader.ReadToEnd();
            }
        }
    }

    public static string GetRecommendations(List<Results> results)
    {
        var recommendations = string.Empty;
        var types = results.Select(x => x.Type).Distinct();
        foreach (var type in types)
        {
            var recommendationsTemplate = GetTemplate($"Resources.{type.Replace("/", ".")}.md");
            recommendations += recommendationsTemplate + Environment.NewLine;
        }
        return recommendations;
    }
}