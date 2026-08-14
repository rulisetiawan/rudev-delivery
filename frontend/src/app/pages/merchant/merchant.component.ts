import { Component, signal, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { TableModule } from 'primeng/table';
import { TagModule } from 'primeng/tag';
import { DialogModule } from 'primeng/dialog';

interface MerchantProduct {
  id: number;
  store_id: number;
  name: string;
  price: number;
  category: string;
  is_available: boolean;
  image_url: string;
}

@Component({
  selector: 'app-merchant-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    CardModule,
    InputTextModule,
    TableModule,
    TagModule,
    DialogModule
  ],
  template: `
    <div class="min-h-screen bg-slate-900 text-slate-100 p-4 md:p-6">
      <!-- HEADER NAVBAR -->
      <div class="flex flex-column md:flex-row justify-content-between align-items-center mb-5 bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
        <div class="flex align-items-center gap-3">
          <span class="text-3xl">🏬</span>
          <div>
            <h1 class="text-2xl font-black m-0 text-emerald-400">PORTAL WARUNG MITRA DESA</h1>
            <p class="text-sm text-slate-400 m-0 mt-1">Kelola stok menu & kelola pesanan toko secara mandiri</p>
          </div>
        </div>
        <div class="flex gap-2 mt-3 md:mt-0">
          <p-button label="+ Tambah Menu Warung" icon="pi pi-plus" styleClass="p-button-emerald border-round-xl" (onClick)="displayAddProductModal.set(true)"></p-button>
          <a href="/#" class="p-button p-button-outlined p-button-success border-round-xl no-underline">
            🌐 Kembali ke Web Utama
          </a>
        </div>
      </div>

      <!-- SUMMARY CARDS -->
      <div class="grid mb-5">
        <div class="col-12 md:col-4">
          <div class="bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
            <span class="text-xs font-bold text-slate-400">TOTAL PRODUK WARUNG</span>
            <h2 class="text-3xl font-black text-white m-0 mt-2">{{ products().length }} Menu</h2>
          </div>
        </div>
        <div class="col-12 md:col-4">
          <div class="bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
            <span class="text-xs font-bold text-slate-400">KOMISI DEDIKASI BUMDES</span>
            <h2 class="text-3xl font-black text-emerald-400 m-0 mt-2">5% (Toko Mitra)</h2>
          </div>
        </div>
        <div class="col-12 md:col-4">
          <div class="bg-slate-800 p-4 border-round-2xl border-1 border-slate-700">
            <span class="text-xs font-bold text-slate-400">STATUS WARUNG</span>
            <h2 class="text-3xl font-black text-emerald-400 m-0 mt-2">🟢 AKTIF BUKA</h2>
          </div>
        </div>
      </div>

      <!-- PRODUCTS TABLE WITH STOK TOGGLE -->
      <div class="p-card p-4 border-round-2xl">
        <h3 class="text-lg font-bold mb-4">🍲 Daftar Menu Makanan & Stok Warung</h3>
        <p-table [value]="products()" [paginator]="true" [rows]="8" styleClass="p-datatable-sm">
          <ng-template pTemplate="header">
            <tr>
              <th>ID</th>
              <th>Nama Makanan/Barang</th>
              <th>Harga (Rp)</th>
              <th>Kategori</th>
              <th>Status Stok</th>
              <th>Aksi Cepat</th>
            </tr>
          </ng-template>
          <ng-template pTemplate="body" let-product>
            <tr>
              <td>{{ product.id }}</td>
              <td class="font-bold">{{ product.name }}</td>
              <td class="text-emerald-400 font-bold">Rp {{ product.price | number:'1.0-0' }}</td>
              <td>{{ product.category }}</td>
              <td>
                <p-tag [value]="product.is_available ? 'Tersedia (Ready)' : 'Habis (Sold Out)'" [severity]="product.is_available ? 'success' : 'danger'"></p-tag>
              </td>
              <td>
                <p-button 
                  [label]="product.is_available ? 'Set Habis' : 'Set Ready'" 
                  [styleClass]="product.is_available ? 'p-button-warning p-button-sm' : 'p-button-success p-button-sm'"
                  (onClick)="toggleStock(product)">
                </p-button>
              </td>
            </tr>
          </ng-template>
        </p-table>
      </div>

      <!-- ADD PRODUCT MODAL FOR MERCHANT -->
      <p-dialog header="🍲 Tambah Menu Warung Baru" [(visible)]="displayAddProductModal" [modal]="true" [style]="{width: '90vw', maxWidth: '450px'}">
        <div class="flex flex-column gap-3 py-2">
          <input type="text" pInputText [(ngModel)]="newProdName" placeholder="Nama Makanan (misal: Nasi Timbel Komplit)" />
          <input type="number" pInputText [(ngModel)]="newProdPrice" placeholder="Harga (Rp)" />
          <input type="text" pInputText [(ngModel)]="newProdCategory" placeholder="Kategori (Makanan/Minuman)" />
          <input type="text" pInputText [(ngModel)]="newProdImg" placeholder="URL Foto Produk / MinIO" />
          <p-button label="💾 Simpan Menu Warung" styleClass="p-button-emerald w-full" (onClick)="saveProduct()"></p-button>
        </div>
      </p-dialog>
    </div>
  `
})
export class MerchantDashboardComponent implements OnInit {
  private http = inject(HttpClient);

  products = signal<MerchantProduct[]>([]);
  displayAddProductModal = signal<boolean>(false);

  newProdName = '';
  newProdPrice = 15000;
  newProdCategory = 'Makanan';
  newProdImg = '';

  ngOnInit() {
    this.fetchMerchantProducts();
  }

  fetchMerchantProducts() {
    this.http.get<MerchantProduct[]>('/api/v1/public/products?store_id=1').subscribe({
      next: (data: MerchantProduct[]) => this.products.set(data || [])
    });
  }

  toggleStock(product: MerchantProduct) {
    const updated = !product.is_available;
    this.http.put<any>('/api/v1/orders/admin/products', {
      id: product.id,
      name: product.name,
      price: product.price,
      category: product.category,
      image_url: product.image_url,
      is_available: updated
    }).subscribe({
      next: () => {
        this.fetchMerchantProducts();
      }
    });
  }

  saveProduct() {
    this.http.post<any>('/api/v1/orders/admin/products', {
      store_id: 1,
      name: this.newProdName,
      price: this.newProdPrice,
      category: this.newProdCategory,
      image_url: this.newProdImg,
      is_available: true
    }).subscribe({
      next: (res: any) => {
        alert(res.message || 'Menu warung berhasil ditambahkan!');
        this.displayAddProductModal.set(false);
        this.fetchMerchantProducts();
      }
    });
  }
}
