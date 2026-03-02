import os

dirs = [
    "svc-query", "svc-alert", "svc-track", "svc-training",
    "svc-fusion-engine", "svc-nato-adapter", "svc-audit",
    "svc-anomaly-detection", "svc-feedback"
]

insert1 = """RUN GRPC_HEALTH_PROBE_VERSION=v0.4.25 && \\
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \\
    chmod +x /bin/grpc_health_probe

"""

insert2 = "COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe\n"

for d in dirs:
    fp = os.path.join(d, "Dockerfile")
    with open(fp, "r") as f:
        content = f.read()

    if "grpc_health_probe" in content:
        continue

    # insert before `# Build the binary`
    content = content.replace("# Build the binary", insert1 + "# Build the binary")

    # insert after `FROM gcr.io/distroless/static-debian12:nonroot`
    distroless = "FROM gcr.io/distroless/static-debian12:nonroot\n"
    content = content.replace(distroless, distroless + "\n" + insert2)

    with open(fp, "w") as f:
        f.write(content)

print("Done")
