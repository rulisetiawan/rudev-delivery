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
              <h2 class="text-lg font-black m-0 text-emerald-400">ADMIN DESA</h2>
              <p class="text-xs text-slate-400 m-0">Pengelola Platform BUMDes</p>
            </div>
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
            <button class="p-button-text p-button-plain text-left w-full border-round-xl p-3 text-white font-bold" [class.bg-emerald-800]="activeTab() === 'tariffs'" (click)="activeTab.set('tariffs')">
              ⚙️ Pengaturan Tarif
            </button>
          </nav>
        </div>
        <div class="text-xs text-slate-500 text-center">
          Desa Delivery Admin Dashboard v1.0
        </div>
      </aside>

      <!-- MAIN CONTENT AREA -->
      <main class="flex-1 p-4 md:p-6 overflow-y-auto">
        <!-- TOP HEADER BAR -->
        <div class="flex justify-content-between align-items-center mb-5 bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
          <div>
            <h1 class="text-2xl font-black m-0">Dashboard Management Admin</h1>
            <p class="text-sm text-slate-400 m-0 mt-1">Kelola data warung, katalog produk, driver, dan tarif desa</p>
          </div>
          <a href="/#" class="p-button p-button-outlined p-button-success border-round-xl no-underline">
            🌐 Lihat Web Customer PWA
          </a>
        </div>

        <!-- TAB 1: STORES MANAGEMENT -->
        @if (activeTab() === 'stores') {
          <div class="p-card p-4 border-round-2xl mb-4">
            <div class="flex justify-content-between align-items-center mb-4">
              <h3 class="text-lg font-bold m-0">🏪 Daftar Warung & Toko Desa</h3>
              <p-button label="+ Tambah Warung Baru" icon="pi pi-plus" styleClass="p-button-emerald border-round-xl" (onClick)="displayAddStoreModal.set(true)"></p-button>
            </div>
            <p-table [value]="stores()" [paginator]="true" [rows]="5" styleClass="p-datatable-sm">
              <ng-template pTemplate="header">
                <tr>
                  <th>ID</th>
                  <th>Nama Warung</th>
                  <th>Alamat RT/RW</th>
                  <th>Status Mitra</th>
                  <th>Aksi</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-store>
                <tr>
                  <td>{{ store.id }}</td>
                  <td class="font-bold">{{ store.store_name }}</td>
                  <td>{{ store.address_text }}</td>
                  <td><p-tag [value]="store.is_partner ? 'Mitra Resmi' : 'Jasa Titip'" [severity]="store.is_partner ? 'success' : 'info'"></p-tag></td>
                  <td>
                    <p-button icon="pi pi-pencil" styleClass="p-button-text p-button-sm"></p-button>
                  </td>
                </tr>
              </ng-template>
            </p-table>
          </div>
        }

        <!-- TAB 2: PRODUCTS MANAGEMENT -->
        @if (activeTab() === 'products') {
          <div class="p-card p-4 border-round-2xl mb-4">
            <div class="flex justify-content-between align-items-center mb-4">
              <h3 class="text-lg font-bold m-0">🍲 Katalog Produk Makanan & Sembako</h3>
              <p-button label="+ Tambah Produk Katalog" icon="pi pi-plus" styleClass="p-button-emerald border-round-xl" (onClick)="displayAddProductModal.set(true)"></p-button>
            </div>
            <p-table [value]="products()" [paginator]="true" [rows]="8" styleClass="p-datatable-sm">
              <ng-template pTemplate="header">
                <tr>
                  <th>ID</th>
                  <th>Nama Produk</th>
                  <th>ID Warung</th>
                  <th>Harga (Rp)</th>
                  <th>Kategori</th>
                  <th>Status</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-product>
                <tr>
                  <td>{{ product.id }}</td>
                  <td class="font-bold">{{ product.name }}</td>
                  <td>Warung #{{ product.store_id }}</td>
                  <td class="text-emerald-400 font-bold">Rp {{ product.price | number:'1.0-0' }}</td>
                  <td>{{ product.category }}</td>
                  <td><p-tag [value]="product.is_available ? 'Tersedia' : 'Habis'" [severity]="product.is_available ? 'success' : 'danger'"></p-tag></td>
                </tr>
              </ng-template>
            </p-table>
          </div>
        }

        <!-- TAB 3: COURIERS MANAGEMENT & LOANS -->
        @if (activeTab() === 'couriers') {
          <div class="p-card p-4 border-round-2xl mb-4">
            <div class="flex justify-content-between align-items-center mb-4">
              <h3 class="text-lg font-bold m-0">🛵 Manajemen Driver Desa & Modal Talangan</h3>
              <div class="flex gap-2">
                <p-button label="+ Daftarkan Driver" icon="pi pi-user-plus" styleClass="p-button-success border-round-xl" (onClick)="displayAddDriverModal.set(true)"></p-button>
                <p-button label="💵 Modal Talangan Kas" icon="pi pi-wallet" styleClass="p-button-emerald border-round-xl" (onClick)="displayAddLoanModal.set(true)"></p-button>
              </div>
            </div>
            <p-table [value]="couriers()" [paginator]="true" [rows]="5" styleClass="p-datatable-sm">
              <ng-template pTemplate="header">
                <tr>
                  <th>ID</th>
                  <th>Nama Driver</th>
                  <th>No. WhatsApp</th>
                  <th>Plat Motor</th>
                  <th>Status Kerja</th>
                  <th>Total Antar</th>
                </tr>
              </ng-template>
              <ng-template pTemplate="body" let-courier>
                <tr>
                  <td>{{ courier.id }}</td>
                  <td class="font-bold">{{ courier.name }}</td>
                  <td>{{ courier.phone_number }}</td>
                  <td><span class="bg-slate-700 px-2 py-1 border-round font-mono text-xs">{{ courier.vehicle_plate }}</span></td>
                  <td><p-tag [value]="courier.status" severity="success"></p-tag></td>
                  <td>{{ courier.total_deliveries }} Pesanan</td>
                </tr>
              </ng-template>
            </p-table>
          </div>
        }

        <!-- TAB 4: TARIFF SETTINGS -->
        @if (activeTab() === 'tariffs') {
          <div class="p-card p-4 border-round-2xl mb-4 max-w-30rem">
            <h3 class="text-lg font-bold mb-3">⚙️ Pengaturan Tarif Pengantaran Dynamic</h3>
            <div class="flex flex-column gap-3">
              <div>
                <label class="block text-xs font-bold text-slate-400 mb-1">Tarif Dasar Pengantaran (Rp)</label>
                <input type="number" pInputText class="w-full" [(ngModel)]="tariffBase" />
              </div>
              <div>
                <label class="block text-xs font-bold text-slate-400 mb-1">Biaya Per Toko Tambahan (Rp)</label>
                <input type="number" pInputText class="w-full" [(ngModel)]="tariffExtra" />
              </div>
              <p-button label="💾 Simpan Pengaturan Tarif" styleClass="p-button-emerald w-full border-round-xl mt-2" (onClick)="saveTariffSettings()"></p-button>
            </div>
          </div>
        }
      </main>

      <!-- MODALS FOR ADMIN ACTIONS -->
      <p-dialog header="➕ Tambah Warung Baru" [(visible)]="displayAddStoreModal" [modal]="true" [style]="{width: '90vw', maxWidth: '450px'}">
        <div class="flex flex-column gap-3 py-2">
          <input type="text" pInputText [(ngModel)]="newStoreName" placeholder="Nama Warung (misal: Bakso Pak No)" />
          <input type="text" pInputText [(ngModel)]="newStoreAddress" placeholder="Alamat RT/RW Warung" />
          <input type="text" pInputText [(ngModel)]="newStoreImg" placeholder="URL Foto Warung / MinIO" />
          <p-button label="💾 Simpan Warung" styleClass="p-button-emerald w-full" (onClick)="saveStore()"></p-button>
        </div>
      </p-dialog>

      <p-dialog header="➕ Tambah Produk Katalog Baru" [(visible)]="displayAddProductModal" [modal]="true" [style]="{width: '90vw', maxWidth: '450px'}">
        <div class="flex flex-column gap-3 py-2">
          <input type="number" pInputText [(ngModel)]="newProdStoreID" placeholder="ID Warung (misal: 1)" />
          <input type="text" pInputText [(ngModel)]="newProdName" placeholder="Nama Produk (misal: Mie Ayam Urat)" />
          <input type="number" pInputText [(ngModel)]="newProdPrice" placeholder="Harga (Rp)" />
          <input type="text" pInputText [(ngModel)]="newProdCategory" placeholder="Kategori Makanan/Minuman" />
          <input type="text" pInputText [(ngModel)]="newProdImg" placeholder="URL Foto Produk / MinIO" />
          <p-button label="💾 Simpan Produk" styleClass="p-button-emerald w-full" (onClick)="saveProduct()"></p-button>
        </div>
      </p-dialog>

      <p-dialog header="➕ Pendaftaran Driver Baru" [(visible)]="displayAddDriverModal" [modal]="true" [style]="{width: '90vw', maxWidth: '450px'}">
        <div class="flex flex-column gap-3 py-2">
          <input type="text" pInputText [(ngModel)]="newDriverName" placeholder="Nama Driver" />
          <input type="text" pInputText [(ngModel)]="newDriverPhone" placeholder="No. WA Driver (628xxx)" />
          <input type="text" pInputText [(ngModel)]="newDriverPlate" placeholder="Plat Motor (misal: Z 4589 AB)" />
          <p-button label="+ Daftarkan Driver" styleClass="p-button-success w-full" (onClick)="saveDriver()"></p-button>
        </div>
      </p-dialog>

      <p-dialog header="💵 Modal Talangan Kas Admin Pagi" [(visible)]="displayAddLoanModal" [modal]="true" [style]="{width: '90vw', maxWidth: '450px'}">
        <div class="flex flex-column gap-3 py-2">
          <input type="number" pInputText [(ngModel)]="loanDriverID" placeholder="ID Driver (misal: 1)" />
          <input type="number" pInputText [(ngModel)]="loanAmount" placeholder="Nominal Modal Kas (misal: 300000)" />
          <p-button label="💸 Berikan Kas Talangan Pagi" styleClass="p-button-emerald w-full" (onClick)="saveLoan()"></p-button>
        </div>
      </p-dialog>
    </div>
  `
})
export class AdminDashboardComponent implements OnInit {
  private http = inject(HttpClient);

  activeTab = signal<'stores' | 'products' | 'couriers' | 'tariffs'>('stores');
  stores = signal<any[]>([]);
  products = signal<any[]>([]);
  couriers = signal<any[]>([]);

  displayAddStoreModal = signal<boolean>(false);
  displayAddProductModal = signal<boolean>(false);
  displayAddDriverModal = signal<boolean>(false);
  displayAddLoanModal = signal<boolean>(false);

  newStoreName = '';
  newStoreAddress = '';
  newStoreImg = '';

  newProdStoreID = 1;
  newProdName = '';
  newProdPrice = 15000;
  newProdCategory = 'Makanan';
  newProdImg = '';

  newDriverName = '';
  newDriverPhone = '';
  newDriverPlate = '';

  loanDriverID = 1;
  loanAmount = 300000;

  tariffBase = 10000;
  tariffExtra = 2000;

  ngOnInit() {
    this.fetchStores();
    this.fetchProducts();
    this.fetchCouriers();
    this.fetchTariffs();
  }

  fetchStores() {
    this.http.get<any[]>('/api/v1/orders/admin/stores').subscribe({
      next: (data: any[]) => this.stores.set(data || [])
    });
  }

  fetchProducts() {
    this.http.get<any[]>('/api/v1/orders/admin/products').subscribe({
      next: (data: any[]) => this.products.set(data || [])
    });
  }

  fetchCouriers() {
    this.http.get<any[]>('/api/v1/orders/admin/couriers').subscribe({
      next: (data: any[]) => this.couriers.set(data || [])
    });
  }

  fetchTariffs() {
    this.http.get<any>('/api/v1/public/settings/tariff').subscribe({
      next: (data: any) => {
        if (data && data.base_delivery_fee) {
          this.tariffBase = data.base_delivery_fee;
          this.tariffExtra = data.per_extra_store_fee;
        }
      }
    });
  }

  saveStore() {
    this.http.post<any>('/api/v1/orders/admin/stores', {
      store_name: this.newStoreName,
      address_text: this.newStoreAddress,
      image_url: this.newStoreImg,
      is_partner: false
    }).subscribe({
      next: (res: any) => {
        alert(res.message || 'Warung berhasil disimpan!');
        this.displayAddStoreModal.set(false);
        this.fetchStores();
      }
    });
  }

  saveProduct() {
    this.http.post<any>('/api/v1/orders/admin/products', {
      store_id: this.newProdStoreID,
      name: this.newProdName,
      price: this.newProdPrice,
      category: this.newProdCategory,
      image_url: this.newProdImg,
      is_available: true
    }).subscribe({
      next: (res: any) => {
        alert(res.message || 'Produk berhasil disimpan!');
        this.displayAddProductModal.set(false);
        this.fetchProducts();
      }
    });
  }

  saveDriver() {
    this.http.post<any>('/api/v1/orders/admin/couriers', {
      name: this.newDriverName,
      phone_number: this.newDriverPhone,
      vehicle_plate: this.newDriverPlate
    }).subscribe({
      next: (res: any) => {
        alert(res.message || 'Driver berhasil didaftarkan!');
        this.displayAddDriverModal.set(false);
        this.fetchCouriers();
      }
    });
  }

  saveLoan() {
    this.http.post<any>('/api/v1/orders/admin/couriers/loan', {
      courier_id: this.loanDriverID,
      amount: this.loanAmount
    }).subscribe({
      next: (res: any) => {
        alert(res.message || 'Modal kas admin berhasil dipinjamkan!');
        this.displayAddLoanModal.set(false);
      }
    });
  }

  saveTariffSettings() {
    this.http.post<any>('/api/v1/orders/admin/settings/tariff', {
      base_delivery_fee: this.tariffBase,
      per_extra_store_fee: this.tariffExtra,
      min_order_amount: 0
    }).subscribe({
      next: (res: any) => alert(res.message || 'Tarif berhasil diperbarui!')
    });
  }
}
