package youtube

type Video struct {
    ID      string
    Title   string
    Author  string
    Length  int
    Formats []Format
}

type Format struct {
    Itag         int
    QualityLabel string
    MimeType     string
    Bitrate      int
    URL          string
    HasAudio     bool
    HasVideo     bool
}