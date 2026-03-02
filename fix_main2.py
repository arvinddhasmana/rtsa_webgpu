import os
import subprocess

grpc_services = ["svc-query", "svc-alert", "svc-track", "svc-audit", "svc-feedback"]
no_grpc_services = ["svc-training", "svc-fusion-engine", "svc-anomaly-detection", "svc-nato-adapter"]

def fix_file(svc, fp):
    with open(fp, "r") as f:
        content = f.read()

    # REMOVE old bad inject
    old_block_grpc = """	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)"""

    old_block_nogrpc = """	// --- Dummy gRPC Health Server ---
	grpcLis, _ := net.Listen("tcp", ":50051")
	grpcServer := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	go grpcServer.Serve(grpcLis)
	defer grpcServer.GracefulStop()"""

    content = content.replace(old_block_grpc, "")
    content = content.replace(old_block_nogrpc, "")

    if "grpchealth.NewServer" in content:
        return

    # Now add the correct block
    if svc in grpc_services:
        target = "grpc.NewServer"
        idx = content.find(target)
        if idx != -1:
            eol = content.find("\n", idx)
            insert_block = """
	grpchealth_server := grpchealth.NewServer()
	grpchealth_server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, grpchealth_server)
"""
            content = content[:eol+1] + insert_block + content[eol+1:]
    else:
        idx = content.find("logger")
        if idx == -1:
            idx = content.find("cancel :=")
        eol = content.find("\n", idx)
        if eol != -1:
            insert_block = """
	// --- Dummy gRPC Health Server ---
	grpcLis, _ := stdnet.Listen("tcp", ":50051")
	grpcServerDummy := grpc.NewServer()
	grpchealth_server := grpchealth.NewServer()
	grpchealth_server.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServerDummy, grpchealth_server)
	go grpcServerDummy.Serve(grpcLis)
	defer grpcServerDummy.GracefulStop()
"""
            content = content[:eol+1] + insert_block + content[eol+1:]

    # Add explicit imports if they are needed, but goimports might rename them
    # But wait, we used `stdnet` instead of `net` to avoid shadowing.
    # Let's just use `net` since it's standard go library.
    content = content.replace("stdnet.Listen", "net.Listen")

    # Let's forcefully add the grpchealth import alias so goimports doesn't trip on it
    if "grpchealth" not in content:
        import_block = """
import (
    grpchealth "google.golang.org/grpc/health"
    "google.golang.org/grpc/health/grpc_health_v1"
)
"""
        # wait! we can't just inject import blindly if there's already an import block.
        # goimports usually handles packages automatically EXCEPT when there's an alias.
        # But if we just write the code without alias, can goimports resolve it?
        pass

    with open(fp, "w") as f:
        f.write(content)

    # We must replace `grpchealth "google.golang.org/grpc/health"` manually at the top of file
    import_line = 'grpchealth "google.golang.org/grpc/health"'
    with open(fp, "r") as f:
        lines = f.readlines()
    for i, line in enumerate(lines):
        if '"google.golang.org/grpc/health/grpc_health_v1"' in line:
            lines.insert(i, import_line + "\n")
            break
        elif '"google.golang.org/grpc"' in line:
            lines.insert(i, import_line + "\n")
            lines.insert(i+1, '\t"google.golang.org/grpc/health/grpc_health_v1"\n')
            break

    with open(fp, "w") as f:
        f.writelines(lines)

    subprocess.run(["goimports", "-w", fp])

def process_all():
    all_svcs = grpc_services + no_grpc_services
    for svc in all_svcs:
        fp = os.path.join(svc, "cmd", svc.split("-")[-1] if svc not in ["svc-query", "svc-alert", "svc-track", "svc-audit", "svc-feedback"] else svc.split("-")[1], "main.go")
        if not os.path.exists(fp):
            for root, dirs, files in os.walk(os.path.join(svc, "cmd")):
                for file in files:
                    if file == "main.go":
                        fp = os.path.join(root, file)
                        break
        fix_file(svc, fp)

process_all()
