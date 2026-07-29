#!/bin/bash
set -e

SERVICES=("svc-ais-ingestion" "svc-cyber-ingestion" "svc-elint-ingestion" "svc-isr-ingestion")
PREFIXES=("AIS" "CYBER" "ELINT" "ISR")

for i in "${!SERVICES[@]}"; do
    SVC="${SERVICES[$i]}"
    PREFIX="${PREFIXES[$i]}"
    CONFIG_FILE="$SVC/internal/config/config.go"

    if [ ! -f "$CONFIG_FILE" ]; then
        echo "Skipping $SVC, config not found"
        continue
    fi

    # 1. Add extra imports if needed
    if ! grep -q "\"github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1\"" "$CONFIG_FILE"; then
		sed -i '/pkgconfig "github.com\/arvinddhasmana\/rtsa_webgpu\/pkg\/config"/i \	commonv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/common/v1"\n	ingestionv1 "github.com/arvinddhasmana/rtsa_webgpu/gen/go/rtsa/ingestion/v1"' "$CONFIG_FILE"
    fi

    # 2. Add Coverage field to Config struct
    if ! grep -q "Coverage \*ingestionv1.SensorCoverage" "$CONFIG_FILE"; then
        sed -i '/^}/ s/^/\t\/\/ Optional static sensor coverage geometry.\n\tCoverage \*ingestionv1.SensorCoverage\n/' "$CONFIG_FILE"
    fi

    # 3. Add Coverage parsing logic before "return cfg, nil"
    if ! grep -q "// Parse coverage" "$CONFIG_FILE"; then
        sed -i "/return cfg, nil/i \\
	// Parse coverage\\
	if rangeNmStr := os.Getenv(\"RTSA_${PREFIX}_RANGE_NM\"); rangeNmStr != \"\" {\\
		if rangeNm, err := strconv.ParseFloat(rangeNmStr, 64); err == nil {\\
			if cfg.Coverage == nil {\\
				cfg.Coverage = &ingestionv1.SensorCoverage{}\\
			}\\
			cfg.Coverage.RangeNm = &rangeNm\\
		}\\
	}\\
	if bStartStr := os.Getenv(\"RTSA_${PREFIX}_BEARING_START\"); bStartStr != \"\" {\\
		if bStart, err := strconv.ParseFloat(bStartStr, 64); err == nil {\\
			if cfg.Coverage == nil {\\
				cfg.Coverage = &ingestionv1.SensorCoverage{}\\
			}\\
			cfg.Coverage.BearingStartDegrees = &bStart\\
		}\\
	}\\
	if bEndStr := os.Getenv(\"RTSA_${PREFIX}_BEARING_END\"); bEndStr != \"\" {\\
		if bEnd, err := strconv.ParseFloat(bEndStr, 64); err == nil {\\
			if cfg.Coverage == nil {\\
				cfg.Coverage = &ingestionv1.SensorCoverage{}\\
			}\\
			cfg.Coverage.BearingEndDegrees = &bEnd\\
		}\\
	}\\
	if latStr := os.Getenv(\"RTSA_${PREFIX}_LAT\"); latStr != \"\" {\\
		if lonStr := os.Getenv(\"RTSA_${PREFIX}_LON\"); lonStr != \"\" {\\
			lat, er1 := strconv.ParseFloat(latStr, 64)\\
			lon, er2 := strconv.ParseFloat(lonStr, 64)\\
			if er1 == nil && er2 == nil {\\
				if cfg.Coverage == nil {\\
					cfg.Coverage = &ingestionv1.SensorCoverage{}\\
				}\\
				cfg.Coverage.SensorPosition = &commonv1.Position{\\
					Latitude:  lat,\\
					Longitude: lon,\\
				}\\
			}\\
		}\\
	}\\
" "$CONFIG_FILE"
    fi

    # 4. Make sure parseFloat exists since some services (like Cyber) might not have it
    if ! grep -q "func parseFloat" "$CONFIG_FILE"; then
        cat << 'SUBEOF' >> "$CONFIG_FILE"

func parseFloat(key string, defaultVal float64) float64 {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(val), 64)
	if err != nil {
		return defaultVal
	}
	return f
}
SUBEOF
    fi

    # Format the file
    gofmt -w "$CONFIG_FILE"
    echo "Updated $SVC"
done
