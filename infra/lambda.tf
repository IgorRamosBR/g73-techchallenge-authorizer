resource "aws_lambda_function" "authorizer_lambda" {
  s3_bucket = "g73-techchallenge-lambda-build"
  s3_key = "authorizer.zip"
  function_name = "api-authorizer"
  role = aws_iam_role.lambda_role.arn
  handler = "bin/main"
  runtime = "provided.al2023"
  environment {
    variables = {
      DB_HOST = "g73-techchallenge-db.cxokeewukuer.us-east-1.rds.amazonaws.com"
      DB_PORT = "5432"
      DB_USER = "g73_admin_user"
      DB_PASSWORD = "UV6RetyeibtF"
      DB_NAME = "techchallengedb"
    }
  }
}

resource "aws_iam_role" "lambda_role" {
  name = "lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17",
    Statement = [
    {
      Action = "sts:AssumeRole",
      Effect = "Allow",
      Principal = {
        Service = "lambda.amazonaws.com"
      }
    }
  ]
})
}

resource "aws_iam_role_policy_attachment" "lambda_basic" {
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
  role = aws_iam_role.lambda_role.name
}

resource "aws_lambda_permission" "apigw_lambda" {
  statement_id = "AllowExecutionFromAPIGateway"
  action = "lambda:InvokeFunction"
  function_name = aws_lambda_function.authorizer_lambda.function_name
  principal = "apigateway.amazonaws.com"

  source_arn = "${aws_api_gateway_rest_api.api_authorizer.execution_arn}/*/*/*"
}

