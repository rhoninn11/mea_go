from grpc_tools import protoc
import os

# add to os path
def proto_gen(in_dir, out_dir):
    ls = os.listdir(in_dir)
    protos = []

    for file in ls:
        if file.endswith(".proto"):
            proto = os.path.join(in_dir, file)
            protos.append(proto)
    print(f'+++ files: {len(ls)}, proto: {len(protos)}')
    
    os.makedirs(out_dir, exist_ok=True)
    for proto in protos:
        protoc.main((
            "",
            f"-I{in_dir}",
            f"--python_out={out_dir}",
            f"--pyi_out={out_dir}",
            f"--grpc_python_out={out_dir}",
            proto,
        ))


def script():
    # relative to root
    IN_DIR="./proto"
    OUT_DIR="./src/mea_gen_d"
    proto_gen(IN_DIR, OUT_DIR)
    print("+++ proto sources regenerated")

script()