import { Component, signal, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { TableModule } from 'primeng/table';
import { TagModule } from 'primeng/tag';
import { TabViewModule } from 'primeng/tabview';
import { DialogModule } from 'primeng/dialog';

interface Store {
  id: number;
  store_name: string;
  slug: string;
  address_text: string;
  image_url: string;
  is_partner: boolean;
  is_active: boolean;
}

interface Product {
  id: number;
  store_id: number;
  name: string;
  price: number;
  category: string;
  is_available: boolean;
  image_url: string;
}

interface Courier {
  id: number;
  user_id: number;
  name: string;
  phone_number: string;
  vehicle_plate: string;
  vehicle_type: string;
  is_online: boolean;
  current_status: string;
  total_deliveries: number;
  rating: number;
}

interface Village {
  id: number;
  name: string;
  district_name: string;
}

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    CardModule,
    InputTextModule,
    TableModule,
    TagModule,
    TabViewModule,
    DialogModule
  ],
  template: `
    <div class="min-h-screen bg-slate-900 text-slate-100 flex flex-column md:flex-row">
      <!-- SIDEBAR ADMIN -->
      <aside class="w-full md:w-16rem bg-slate-800 p-4 border-right-1 border-slate-700 flex flex-column justify-content-between">
        <div>
          <div class="flex align-items-center gap-2 mb-4">
            <span class="text-2xl">⚙️</span>
            <div>
              <h2 class="text-lg font-black m-0 text-emerald-400">ADMIN BUMDES</h2>
              <p class="text-xs text-slate-400 m-0">Cluster Kecamatan Multi-Desa</p>
            </div>
          </div>

          <!-- MULTI-VILLAGE SELECTOR (PHASE 3) -->
          <div class="mb-4 bg-slate-900 p-3 border-round-xl">
            <label class="block text-xs font-bold text-slate-400 mb-1">🏢 Pilih Cluster Desa:</label>
            <select class="w-full bg-slate-800 text-white p-2 border-round-lg border-1 border-slate-700" (change)="onVillageChange($event)">
              <option value="0">🌐 Semua Desa (Kecamatan)</option>
              @for (v of villages(); track v.id) {
                <option [value]="v.id">{{ v.name }}</option>
              }
            </select>
          </div>

          <hr class="border-slate-700 mb-4" />
          <nav class="flex flex-column gap-2">
            <button class="p-button-text p-button-plain text-left w-full border-round-xl p-3 text-white font-bold" [class.bg-emerald-800]="activeTab() === 'stores'" (click)="activeTab.set('stores')">
              🏪 Management Warung
            </button>
            <button class="p-button-text p-button-plain text-left w-full border-round-xl p-3 text-white font-bold" [class.bg-emerald-800]="activeTab() === 'products'" (click)="activeTab.set('products')">
              🍲 Management Katalog
            </button>
            <button class="p-button-text p-button-plain text-left w-full border-round-xl p-3 text-white font-bold" [class.bg-emerald-800]="activeTab() === 'couriers'" (click)="activeTab.set('couriers')">
              🛵 Management Driver
            </button>
            <button class="p-button-text p-button-plain text-left w-full border-round-xl p-3 text-white font-bold" [class.bg-emerald-800]="activeTab() === 'ratings'" (click)="activeTab.set('ratings')">
              ⭐ Ratings & Bonus Driver
            </button>
            <button class="p-button-text p-button-plain text-left w-full border-round-xl p-3 text-white font-bold" [class.bg-emerald-800]="activeTab() === 'analytics'" (click)="activeTab.set('analytics')">
              📊 Metabase Analytics BI
            </button>
            <button class="p-button-text p-button-plain text-left w-full border-round-xl p-3 text-white font-bold" [class.bg-emerald-800]="activeTab() === 'tariffs'" (click)="activeTab.set('tariffs')">
              ⚙️ Pengaturan Tarif
            </button>
          </nav>
        </div>
        <div class="text-xs text-slate-500 text-center">
          Desa Delivery Cluster Admin v3.0
        </div>
      </aside>

      <!-- MAIN CONTENT AREA -->
      <main class="flex-1 p-4 md:p-6 overflow-y-auto">
        <!-- TOP HEADER NAVBAR -->
        <div class="flex flex-column md:flex-row justify-content-between align-items-center mb-5 bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
          <div>
            <h1 class="text-2xl font-black m-0 text-white">SYSTEM CONTROL & MANAGEMENT</h1>
            <p class="text-sm text-slate-400 m-0 mt-1">Kelola data warung, katalog produk, kurir, dan analitis BUMDes</p>
          </div>
          <div class="flex gap-2 mt-3 md:mt-0 flex-wrap">
            <a href="/api/v1/orders/admin/reports/csv" download class="p-button p-button-success border-round-xl no-underline font-bold">
              📥 Export Laporan CSV BUMDes
            </a>
            <p-button label="⚖️ Daily Settlement Harian" styleClass="p-button-warning border-round-xl" (onClick)="runDailySettlement()"></p-button>
            <a href="http://localhost:3005" target="_blank" class="p-button p-button-outlined p-button-info border-round-xl no-underline">
              📊 Metabase BI
            </a>
            <a href="http://localhost:9002" target="_blank" class="p-button p-button-outlined p-button-warning border-round-xl no-underline">
              🛡️ SonarQube SAST
            </a>
          </div>
        </div>

        <!-- TAB 1: MANAGEMENT WARUNG DESA -->
        @if (activeTab() === 'stores') {
          <div class="p-card p-4 border-round-2xl">
            <div class="flex justify-content-between align-items-center mb-4">
              <h3 class="text-lg font-bold m-0">🏪 Daftar Toko & Warung Desa</h3>
              <p-button label="+ Tambah Warung Baru" icon="pi pi-plus" styleClass="p-button-emerald border-round-xl" (onClick)="displayStoreModal.set(true)"></p-button>
            </div>
            <p-table [value]="stores()" [paginator]="true" [rows]="5" styleClass="p-datatable-sm">
              <ng-template pTemplate="header">
                <tr>
                  <th>ID</th>
                  <th>Nama Warung</th>
                  <th>Slug URL</th>
                  <th>Alamat</th>
                  <th>Tipe Kemitraan</th>
                  <th>Status</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-store>
                <tr>
                  <td>{{ store.id }}</td>
                  <td class="font-bold">{{ store.store_name }}</td>
                  <td>{{ store.slug }}</td>
                  <td>{{ store.address_text }}</td>
                  <td>
                    <p-tag [value]="store.is_partner ? '🟢 Toko Mitra' : '🟡 Jasa Titip'" [severity]="store.is_partner ? 'success' : 'warning'"></p-tag>
                  </td>
                  <td>
                    <p-tag [value]="store.is_active ? 'Aktif' : 'Buka'" severity="success"></p-tag>
                  </td>
                </tr>
              </ng-template>
            </p-table>
          </div>
        }

        <!-- TAB 2: MANAGEMENT KATALOG PRODUK -->
        @if (activeTab() === 'products') {
          <div class="p-card p-4 border-round-2xl">
            <div class="flex justify-content-between align-items-center mb-4">
              <h3 class="text-lg font-bold m-0">🍲 Katalog Menu & Barang Warung</h3>
              <p-button label="+ Tambah Produk Baru" icon="pi pi-plus" styleClass="p-button-emerald border-round-xl" (onClick)="displayProductModal.set(true)"></p-button>
            </div>
            <p-table [value]="products()" [paginator]="true" [rows]="8" styleClass="p-datatable-sm">
              <ng-template pTemplate="header">
                <tr>
                  <th>ID</th>
                  <th>Nama Makanan / Barang</th>
                  <th>Harga (Rp)</th>
                  <th>Kategori</th>
                  <th>Ketersediaan Stok</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-prod>
                <tr>
                  <td>{{ prod.id }}</td>
                  <td class="font-bold">{{ prod.name }}</td>
                  <td class="text-emerald-400 font-bold">Rp {{ prod.price | number:'1.0-0' }}</td>
                  <td>{{ prod.category }}</td>
                  <td>
                    <p-tag [value]="prod.is_available ? 'Ready' : 'Habis'" [severity]="prod.is_available ? 'success' : 'danger'"></p-tag>
                  </td>
                </tr>
              </ng-template>
            </p-table>
          </div>
        }

        <!-- TAB 3: MANAGEMENT DRIVER & PINJAMAN MODAL TALANGAN -->
        @if (activeTab() === 'couriers') {
          <div class="p-card p-4 border-round-2xl mb-4">
            <div class="flex justify-content-between align-items-center mb-4">
              <h3 class="text-lg font-bold m-0">🛵 Daftar Kurir & Driver Desa</h3>
              <p-button label="+ Registrasi Driver Baru" icon="pi pi-user-plus" styleClass="p-button-emerald border-round-xl" (onClick)="displayCourierModal.set(true)"></p-button>
            </div>
            <p-table [value]="couriers()" [paginator]="true" [rows]="5" styleClass="p-datatable-sm">
              <ng-template pTemplate="header">
                <tr>
                  <th>ID</th>
                  <th>Nama Driver</th>
                  <th>No WhatsApp</th>
                  <th>Plat Nomor</th>
                  <th>Rating</th>
                  <th>Status Job</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-cour>
                <tr>
                  <td>{{ cour.id }}</td>
                  <td class="font-bold">{{ cour.name }}</td>
                  <td>{{ cour.phone_number }}</td>
                  <td>{{ cour.vehicle_plate }}</td>
                  <td class="text-yellow-400 font-bold">⭐ {{ cour.rating || 5.0 }}</td>
                  <td>
                    <p-tag [value]="cour.is_online ? '🟢 Online (' + cour.current_status + ')' : '🔴 Offline'" [severity]="cour.is_online ? 'success' : 'danger'"></p-tag>
                  </td>
                </tr>
              </ng-template>
            </p-table>
          </div>
        }

        <!-- TAB 4: RATINGS & DRIVER PERFORMANCE BONUS (PHASE 3) -->
        @if (activeTab() === 'ratings') {
          <div class="p-card p-4 border-round-2xl">
            <h3 class="text-lg font-bold mb-3">⭐ Performa Rating & Bonus Insentif Driver Bintang 5</h3>
            <div class="grid mb-4">
              <div class="col-12 md:col-6">
                <div class="bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
                  <span class="text-xs font-bold text-slate-400">TOTAL DRIVER BINTANG 5</span>
                  <h2 class="text-3xl font-black text-emerald-400 m-0 mt-2">100% Top Performer</h2>
                </div>
              </div>
              <div class="col-12 md:col-6">
                <div class="bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
                  <span class="text-xs font-bold text-slate-400">TOTAL BONUS DISALURKAN</span>
                  <h2 class="text-3xl font-black text-emerald-400 m-0 mt-2">Rp 25.000 / Hari</h2>
                </div>
              </div>
            </div>
            <p-table [value]="couriers()" [paginator]="true" [rows]="5" styleClass="p-datatable-sm">
              <ng-template pTemplate="header">
                <tr>
                  <th>Nama Driver</th>
                  <th>Plat Nomor</th>
                  <th>Total Antaran</th>
                  <th>Rerata Rating</th>
                  <th>Bonus Insentif</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-c>
                <tr>
                  <td class="font-bold">{{ c.name }}</td>
                  <td>{{ c.vehicle_plate }}</td>
                  <td>{{ c.total_deliveries }} Pengantaran</td>
                  <td class="text-yellow-400 font-bold">⭐ {{ c.rating || 5.0 }} / 5.0</td>
                  <td class="text-emerald-400 font-bold">Rp 5.000 (Bonus Bintang 5)</td>
                </tr>
              </ng-template>
            </p-table>
          </div>
        }

        <!-- TAB 5: METABASE ANALYTICS BI DASHBOARD (PHASE 3) -->
        @if (activeTab() === 'analytics') {
          <div class="p-card p-4 border-round-2xl">
            <h3 class="text-lg font-bold mb-3">📊 Dashboard Analitis Multi-Desa (Metabase BI Integration)</h3>
            <p class="text-slate-400 text-sm mb-4">Grafik makanan terlaris per desa, jam sibuk pesanan, dan heatmap kurir:</p>
            <div class="bg-slate-800 p-6 border-round-2xl text-center border-1 border-slate-700">
              <span class="text-4xl">📈</span>
              <h3 class="text-xl font-bold mt-2">Metabase Analytics Connected</h3>
              <p class="text-slate-400 text-sm mb-4">Metabase OLAP Engine siap menampilkan 3 Views Analitis PostgreSQL.</p>
              <a href="http://localhost:3005" target="_blank" class="p-button p-button-emerald border-round-xl no-underline font-bold">
                🌐 Buka Metabase BI Dashboard Live
              </a>
            </div>
          </div>
        }

        <!-- TAB 6: PENGATURAN TARIF DYNAMIC -->
        @if (activeTab() === 'tariffs') {
          <div class="p-card p-4 border-round-2xl">
            <h3 class="text-lg font-bold mb-4">⚙️ Pengaturan Tarif Pengantaran Fleksibel (Dynamic Delivery Tariff)</h3>
            <div class="flex flex-column gap-3 max-w-30rem">
              <div>
                <label class="block text-xs font-bold text-slate-400 mb-1">Tarif Dasar Pengantaran Desa (Rp)</label>
                <input type="number" pInputText [(ngModel)]="baseDeliveryFee" class="w-full" />
              </div>
              <div>
                <label class="block text-xs font-bold text-slate-400 mb-1">Biaya Per Toko Tambahan (Rp)</label>
                <input type="number" pInputText [(ngModel)]="perExtraStoreFee" class="w-full" />
              </div>
              <p-button label="💾 Simpan Pengaturan Tarif" styleClass="p-button-emerald border-round-xl font-bold mt-2" (onClick)="saveTariffSettings()"></p-button>
            </div>
          </div>
        }
      </main>

      <!-- MODALS FOR STORE, PRODUCT, COURIER -->
      <p-dialog header="🏪 Form Tambah Warung Baru" [(visible)]="displayStoreModal" [modal]="true" [style]="{width: '90vw', maxWidth: '450px'}">
        <div class="flex flex-column gap-3 py-2">
          <input type="text" pInputText [(ngModel)]="newStoreName" placeholder="Nama Warung (misal: Warung Nasi Bu Ani)" />
          <input type="text" pInputText [(ngModel)]="newStoreSlug" placeholder="Slug URL (misal: warung-bu-ani)" />
          <input type="text" pInputText [(ngModel)]="newStoreAddress" placeholder="Alamat Singkat" />
          <p-button label="Simpan Warung" styleClass="p-button-emerald w-full" (onClick)="saveStore()"></p-button>
        </div>
      </p-dialog>
    </div>
  `
})
export class AdminDashboardComponent implements OnInit {
  private http = inject(HttpClient);

  activeTab = signal<string>('stores');
  stores = signal<Store[]>([]);
  products = signal<Product[]>([]);
  couriers = signal<Courier[]>([]);
  villages = signal<Village[]>([]);

  displayStoreModal = signal<boolean>(false);
  displayProductModal = signal<boolean>(false);
  displayCourierModal = signal<boolean>(false);

  newStoreName = '';
  newStoreSlug = '';
  newStoreAddress = '';

  baseDeliveryFee = 10000;
  perExtraStoreFee = 2000;

  ngOnInit() {
    this.fetchVillages();
    this.fetchStores();
    this.fetchProducts();
    this.fetchCouriers();
    this.fetchTariffs();
  }

  fetchVillages() {
    this.http.get<Village[]>('/api/v1/public/villages').subscribe({
      next: (data: Village[]) => this.villages.set(data || [])
    });
  }

  onVillageChange(event: any) {
    const vId = event.target.value;
    this.fetchStores(vId);
  }

  fetchStores(villageId: number = 0) {
    const url = villageId > 0 ? `/api/v1/public/stores?village_id=${villageId}` : '/api/v1/public/stores';
    this.http.get<Store[]>(url).subscribe({
      next: (data: Store[]) => this.stores.set(data || [])
    });
  }

  fetchProducts() {
    this.http.get<Product[]>('/api/v1/public/products').subscribe({
      next: (data: Product[]) => this.products.set(data || [])
    });
  }

  fetchCouriers() {
    this.http.get<Courier[]>('/api/v1/orders/admin/couriers').subscribe({
      next: (data: Courier[]) => this.couriers.set(data || [])
    });
  }

  fetchTariffs() {
    this.http.get<any>('/api/v1/public/settings/tariff').subscribe({
      next: (data: any) => {
        if (data) {
          this.baseDeliveryFee = data.base_delivery_fee || 10000;
          this.perExtraStoreFee = data.per_extra_store_fee || 2000;
        }
      }
    });
  }

  saveStore() {
    this.http.post<any>('/api/v1/orders/admin/stores', {
      store_name: this.newStoreName,
      slug: this.newStoreSlug,
      address_text: this.newStoreAddress,
      is_partner: true
    }).subscribe({
      next: () => {
        this.displayStoreModal.set(false);
        this.fetchStores();
      }
    });
  }

  saveTariffSettings() {
    this.http.put<any>('/api/v1/orders/admin/settings/tariff', {
      base_delivery_fee: this.baseDeliveryFee,
      per_extra_store_fee: this.perExtraStoreFee
    }).subscribe({
      next: () => alert('Pengaturan tarif berhasil diperbarui!')
    });
  }

  runDailySettlement() {
    this.http.post<any>('/api/v1/orders/admin/settlement/daily', {}).subscribe({
      next: (res: any) => alert(res.message || 'Settlement harian berhasil diselesaikan!')
    });
  }
}
