connection "aws_dmi_001" {
  plugin                = "aws"
  secret_key            = "aws_dmi_001_secret_key"
  access_key            = "aws_dmi_001_access_key"
  regions               = "- us-east-1\n-us-west-"
}

connection "aws_dmi_002" {
  plugin                = "aws" 
  secret_key            = "aws_dmi_002_secret_key"
  access_key            = "aws_dmi_002_access_key"
  regions               = "- us-east-1\n-us-west-"
}

plugin "foo" {
  source = "aws"
}