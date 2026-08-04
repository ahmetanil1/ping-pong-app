# Ping-Pong Microservices Architecture (Kubernetes & Helm)
### 🇹🇷 Türkçe README

Bu proje, geleneksel monolitik yazılım geliştirme pratiklerinden DevOps standartlarına geçişi somutlaştıran bir Proof of Concept (PoC) çalışmasıdır. Kubernetes ortamında çalışmak üzere tasarlanmış; High Availability (HA), otomasyon, güvenlik ve Horizontal Pod Autoscaling (HPA) ilkelerine dayalı production-ready bir mikroservis mimarisidir. Go diliyle yazılmış ping-service ve pong-service uygulamalarının, persistent data storage (PostgreSQL) ve cache (Redis) katmanlarıyla olan senkron iletişimini kapsar.

🏗️ Mimari ve Sistem Tasarımı
Sistem, Separation of Concerns ve Network Isolation ilkelerine uygun olarak iki temel alana ayrılmıştır:

**🌐 Public Zone:
ping-service (:8080):** Dış dünyadan gelen HTTP isteklerini (GET / ve GET /api/v1/ping) karşılayan ana giriş noktasıdır (API Gateway / Producer). İsteği alır, doğrular ve işlem yapması için asenkron/senkron olarak arka katmana iletir. CPU bound süreçlerini kendi üzerinde tutmaz.

**🔒 Private / Isolated Zone:
pong-service (:8081):** Dış erişime tamamen kapalıdır (Worker / Consumer). Yalnızca ping-service tarafından gönderilen HTTP POST /internal/process isteklerini kabul eder (2s timeout) ve asıl business logic yürütür.

**PostgreSQL DB (:5432):** Sistemin RDBMS katmanıdır. pong-service tarafından gelen log verilerini işler (SQL INSERT INTO request_logs).

**Redis Cache (:6379):** Yüksek hızlı in-memory işlemler için kullanılır. Veritabanı yükünü hafifletmek ve pong-service tarafından son işlem zamanını saklamak (SET last_ping TIMESTAMP) için yapılandırılmıştır.

### 🐳 Docker & Continuous Integration (CI)
Geliştirme süreçlerindeki insan faktörünü ortadan kaldıran, güvenlik odaklı GitHub Actions pipeline kurgusu içermektedir:

**Docker Optimizasyonu:** İmaj boyutlarını küçültmek ve Attack Surface alanını daraltmak için Multi-Stage Build stratejisi uygulanmıştır. CI derleme sürelerini hızlandırmak adına layer caching (type=gha) aktiftir.

**Parallel Testing & Linting:** Her kod push işleminde, servislerin unit tests eşzamanlı olarak koşulur ve Helm manifest dosyaları syntax hatalarına karşı otomatik olarak doğrulanır.

**Security Scan (DevSecOps):** CodeQL entegrasyonu ile (SAST) statik kod analizi yapılarak güvenlik zafiyetleri taranır. CI içerisindeki tüm jobs, Principle of Least Privilege kapsamında çalışır.

### ⚙️ Standartlar
**Parametrik Helm Chart Yapısı:** Tüm Kubernetes manifestleri (Deployments, StatefulSet, Services, ConfigMaps, Secrets, HPA) parametrik hale getirilerek values.yaml üzerinden dinamik ve ortamdan bağımsız olarak yönetilmektedir.

**Horizontal Pod Autoscaling (HPA):** ping-service ve pong-service için metrik tabanlı CPU kullanımı (%70 eşiği) izlenerek replica sayıları yük durumuna göre otomatik olarak 2 ile 3 arasında ölçeklenir.

### Gelişmiş Health Probe Katmanı:

**startupProbe:** PostgreSQL gibi Stateful servislerin recovery süreçlerinde Kubernetes tarafından erken sonlandırılmasını engeller (pg_isready).

**readinessProbe:** Uygulama tamamen hazır olana kadar trafiğin Pod'a yönlendirilmesini engelleyerek Zero-Downtime distribution sağlar.

**livenessProbe:** Deadlock olan veya yanıt vermeyen Pod'ları tespit ederek otomatik yeniden başlatır (Self-Healing).

**Graceful Shutdown:** preStop hook'ları sayesinde rolling update sırasında aktif HTTP bağlantılarının yarım kalması ve veri kaybı engellenmiştir.

**Persistence:** PostgreSQL StatefulSet ve PersistentVolumeClaim (PVC) kullanılarak Pod silinme/çökme senaryolarında veri kaybı engellenmiştir.

**Sıkılaştırılmış Kaynak Güvenliği:** Her Pod için resources.requests / limits değerleri ve non-root securityContext tanımlanmıştır.

### 🛠️ Teknolojiler

**Programming Language: Go (Golang)**

**Containerization: Docker**

**Orchestration: Kubernetes (Minikube / Multi-Node Cluster Topology)**

**Package Manager: Helm v3**

**Databases: PostgreSQL 15 (Relational), Redis (In-Memory Cache)**

### 🚀 Kurulum ve Çalıştırma Adımları
**Ön Gereksinimler: Docker & Minikube, kubectl CLI, Helm v3, Make**

1. Cluster Topology ve Namespace İzolasyonu
Bash
### Cluster konfigürasyonunu ve namespace izolasyonunu başlatın
bash scripts/cluster-setup.sh

2. Helm ile Deploy Etme (Local Automation)
Geliştirici ergonomisini (DX) artırmak için local distributions Makefile üzerinden standartlaştırılmıştır.

Bash
### Helm bağımlılıklarını güncelleyip Minikube üzerine dağıtımı yapın
make deploy

### Manuel Helm dağıtımı için alternatif komut:
helm upgrade --install ping-pong-app ./deployments/helm/ping-pong-charts -n ping-pong

3. Servislere Erişme (Local Port-Forwarding)
Bash
### Ping servisi tünelleme (Public Zone)
kubectl port-forward svc/ping-pong-app-hepapi-case-ping-service 8080:8080 -n ping-pong

### Pong servisi tünelleme (İzole Katman Debug İçin)
kubectl port-forward svc/ping-pong-app-hepapi-case-pong-service 8081:8081 -n ping-pong
🧪 Yük ve Autoscaling (HPA) Testi
HPA mekanizmasının ping ve pong servislerini eşzamanlı olarak 3 replikaya çıkardığını doğrulamak için paralelleştirilmiş yük testi:

Bash
while true; do curl -s http://localhost:8080/ping > /dev/null; done &
Canlı HPA ve Pod ölçeklenmesini izlemek için:

Bash
kubectl get hpa -n ping-pong -w