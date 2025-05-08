class CodacyCliV2 < Formula
  version "1.0.0-main.271.4770bfe"
  url "https://raw.githubusercontent.com/codacy/codacy-cli-v2/1.0.0-main.271.4770bfe/codacy-cli.sh"
  sha256 "def852e43b05871b1a5e14dc8fb32015d89e4f4ed33fc3de8570738e6bd180f2"

  def install
    bin.install "codacy-cli.sh" => "codacy-cli"
  end
end
