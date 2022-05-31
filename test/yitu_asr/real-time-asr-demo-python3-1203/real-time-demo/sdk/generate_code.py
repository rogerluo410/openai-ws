"""Run protoc with the gRPC plugin to generate code"""

from grpc_tools import protoc

protoc.main((
	'',
	'-Iprotobuf',
	'--python_out=.',
    '--grpc_python_out=.',
    'protobuf/real_time_asr.proto',
))