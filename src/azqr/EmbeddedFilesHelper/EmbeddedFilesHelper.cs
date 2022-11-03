namespace azqr;

public static class EmbeddedFilesHelper
{
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