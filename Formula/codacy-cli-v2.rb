class CodacyCliV2 < Formula
  version "1.0.0-main.245.02ac1ee"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/1.0.0-main.245.02ac1ee/codacy-cli.sh"
  sha256 "25310825948b5a337a0eb13ef9cd7d52f87f0be5d99b04bacf24a25c97e391a2"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
