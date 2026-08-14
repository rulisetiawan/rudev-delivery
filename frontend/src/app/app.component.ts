import { Component, signal, computed, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient } from '@angular/common/http';
import { FormsModule } from '@angular/forms';
import { ButtonModule } from 'primeng/button';
import { CardModule } from 'primeng/card';
import { InputTextModule } from 'primeng/inputtext';
import { TagModule } from 'primeng/tag';
import { DialogModule } from 'primeng/dialog';
import { AdminDashboardComponent } from './pages/admin/admin.component';
import { MerchantDashboardComponent } from './pages/merchant/merchant.component';

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

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    ButtonModule,
    CardModule,
    InputTextModule,
    TagModule,
    DialogModule,
    AdminDashboardComponent,
    MerchantDashboardComponent
  ],
  template: `
    @if (isAdminRoute()) {
      <!-- DEDICATED DESKTOP WEB ADMIN DASHBOARD -->
      <app-admin-dashboard></app-admin-dashboard>
    } @else if (isMerchantRoute()) {
      <!-- MERCHANT WARUNG MITRA PORTAL DASHBOARD (PHASE 2) -->
      <app-merchant-dashboard></app-merchant-dashboard>
    } @else {
      <!-- PURE CUSTOMER PWA WEB APP -->
      <div class="min-h-screen surface-ground text-white">
        <!-- HEADER NAVBAR -->
        <header class="bg-gradient-to-r from-emerald-800 to-emerald-600 p-4 border-round-bottom-2xl shadow-4 relative">
          <div class="max-w-30rem mx-auto text-center">
            <h1 class="text-2xl font-black m-0">🛵 PESAN ANTAR DESA</h1>
            <p class="text-xs text-emerald-200 mt-1 m-0">Kuliner & Sembako Desa Siap Antar via WhatsApp / QRIS</p>
          </div>
        </header>

        <div class="max-w-30rem mx-auto p-3">
          <!-- SEARCH BOX -->
          <div class="mb-4">
            <span class="p-input-icon-left w-full">
              <i class="pi pi-search"></i>
              <input type="text" pInputText class="w-full border-round-3xl" placeholder="Cari warung, bakso, nasi goreng..." />
            </span>
          </div>

          <!-- STORES GRID -->
          <div class="flex justify-content-between align-items-center mb-3">
            <span class="font-bold text-lg">🏪 Warung & Toko Desa</span>
            <span class="text-xs text-400">Terdekat</span>
          </div>

          <div class="grid grid-nogutter gap-3 mb-4">
            @for (store of stores(); track store.id) {
              <div class="col-5 p-card p-3 border-round-xl cursor-pointer hover:shadow-4 transition-duration-200" (click)="loadStoreProducts(store)">
                <img [src]="store.image_url || 'https://images.unsplash.com/photo-1555396273-367ea4eb4db5?w=300'" class="w-full border-round-md h-6rem object-cover mb-2" />
                <h3 class="text-sm font-bold m-0 text-white">{{ store.store_name }}</h3>
                <p class="text-xs text-400 m-0 mt-1">{{ store.address_text }}</p>
                <p-tag [value]="store.is_partner ? '🟢 Toko Mitra' : '🟢 Jasa Titip'" [severity]="store.is_partner ? 'success' : 'info'" styleClass="mt-2 text-xs"></p-tag>
              </div>
            } @empty {
              <div class="col-12 text-center text-gray-400 py-3">Memuat warung desa...</div>
            }
          </div>

          <!-- PRODUCTS LIST -->
          <div class="font-bold text-lg mb-3">🍲 Menu Makanan & Barang di {{ selectedStoreName() }}</div>

          <div class="flex flex-column gap-3">
            @for (product of products(); track product.id) {
              <div class="p-card p-3 border-round-xl flex align-items-center gap-3">
                <img [src]="product.image_url || 'https://images.unsplash.com/photo-1603133872878-684f208fb84b?w=200'" class="w-5rem h-5rem border-round-md object-cover" />
                <div class="flex-1">
                  <h4 class="text-base font-bold m-0">{{ product.name }}</h4>
                  <p class="text-emerald-400 font-bold m-0 mt-1">Rp {{ product.price | number:'1.0-0' }}</p>
                </div>
                <p-button label="+ Tambah" icon="pi pi-plus" styleClass="p-button-emerald p-button-sm border-round-lg" (onClick)="addToCart(product)"></p-button>
              </div>
            } @empty {
              <div class="text-center text-gray-400 py-4">Pilih warung di atas untuk melihat menu...</div>
            }
          </div>
        </div>

        <!-- FLOATING CART BAR -->
        @if (cartCount() > 0) {
          <div class="fixed bottom-0 left-50 -translate-x-50 w-11 max-w-30rem bg-gradient-to-r from-emerald-600 to-emerald-800 border-round-top-2xl p-3 shadow-6 flex justify-content-between align-items-center cursor-pointer mb-3" (click)="displayCheckoutModal.set(true)">
            <div>
              <h3 class="text-base font-bold m-0">{{ cartCount() }} Item Belanjaan</h3>
              <p class="text-xs text-emerald-200 m-0">Klik untuk checkout & share location WA / QRIS</p>
            </div>
            <div class="text-lg font-black">Rp {{ cartTotal() | number:'1.0-0' }}</div>
          </div>
        }

        <!-- CHECKOUT MODAL -->
        <p-dialog header="📝 Form Checkout Pesanan" [(visible)]="displayCheckoutModal" [modal]="true" [style]="{width: '95vw', maxWidth: '500px'}">
          <div class="flex flex-column gap-3 p-1">
            <div>
              <label class="block text-xs font-bold text-gray-400 mb-1">Nama Lengkap Pembeli</label>
              <input type="text" pInputText class="w-full" [(ngModel)]="customerName" placeholder="Contoh: Pak Agus" />
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-400 mb-1">Nomor WhatsApp (Format 62)</label>
              <input type="text" pInputText class="w-full" [(ngModel)]="customerPhone" placeholder="Contoh: 6281234567890" />
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-400 mb-1">Alamat / Catatan Patokan Rumah</label>
              <input type="text" pInputText class="w-full" [(ngModel)]="customerAddress" placeholder="Contoh: RT 03/RW 01 Depan Masjid" />
            </div>
            <div>
              <label class="block text-xs font-bold text-gray-400 mb-1">Metode Pembayaran</label>
              <div class="flex gap-2">
                <p-button [label]="paymentMethod === 'COD' ? '✅ COD Tunai' : 'COD Tunai'" [styleClass]="paymentMethod === 'COD' ? 'p-button-emerald w-half' : 'p-button-outlined w-half'" (onClick)="paymentMethod = 'COD'"></p-button>
                <p-button [label]="paymentMethod === 'QRIS' ? '✅ Dynamic QRIS' : 'Dynamic QRIS'" [styleClass]="paymentMethod === 'QRIS' ? 'p-button-emerald w-half' : 'p-button-outlined w-half'" (onClick)="paymentMethod = 'QRIS'"></p-button>
              </div>
            </div>

            <div class="bg-gray-800 p-3 border-round-xl mt-2">
              <div class="flex justify-content-between text-sm mb-1">
                <span>Subtotal Makanan</span>
                <span>Rp {{ cartSubtotal() | number:'1.0-0' }}</span>
              </div>
              <div class="flex justify-content-between text-sm mb-1">
                <span>Ongkos Kirim Desa (Flat)</span>
                <span>Rp {{ baseDeliveryFee() | number:'1.0-0' }}</span>
              </div>
              <div class="flex justify-content-between text-base font-bold text-emerald-400 mt-2">
                <span>Total Wajib Bayar</span>
                <span>Rp {{ cartTotal() | number:'1.0-0' }}</span>
              </div>
            </div>

            @if (qrisUrl()) {
              <div class="text-center p-3 bg-white border-round-xl text-black">
                <h4 class="font-bold m-0 mb-2">Scan QRIS BCA/Mandiri/GoPay/OVO:</h4>
                <img [src]="qrisUrl()" class="w-12rem h-12rem mx-auto" />
                <p class="text-xs text-gray-600 m-0 mt-2">Status bayar akan terverifikasi otomatis!</p>
              </div>
            }

            <p-button [label]="paymentMethod === 'QRIS' ? '💳 Bayar dengan Dynamic QRIS' : '📱 Kirim via WhatsApp WAHA'" icon="pi pi-whatsapp" styleClass="p-button-success w-full p-button-lg border-round-xl font-bold mt-2" (onClick)="submitCheckout()"></p-button>
          </div>
        </p-dialog>
      </div>
    }
  `
})
export class AppComponent implements OnInit {
  private http = inject(HttpClient);

  isAdminRoute = signal<boolean>(false);
  isMerchantRoute = signal<boolean>(false);

  stores = signal<Store[]>([]);
  products = signal<Product[]>([]);
  selectedStoreName = signal<string>('Warung Bu Ani');
  selectedStoreID = signal<number>(1);

  cart = signal<Product[]>([]);
  baseDeliveryFee = signal<number>(10000);
  qrisUrl = signal<string>('');

  cartSubtotal = computed(() => this.cart().reduce((sum: number, item: Product) => sum + item.price, 0));
  cartTotal = computed(() => this.cartSubtotal() + this.baseDeliveryFee());
  cartCount = computed(() => this.cart().length);

  displayCheckoutModal = signal<boolean>(false);

  customerName = '';
  customerPhone = '';
  customerAddress = '';
  customNotes = '';
  paymentMethod = 'COD';

  ngOnInit() {
    this.checkRoute();
    if (!this.isAdminRoute() && !this.isMerchantRoute()) {
      this.fetchStores();
      this.fetchTariffs();
    }
  }

  checkRoute() {
    const path = window.location.pathname;
    const hash = window.location.hash;
    if (path.includes('/admin') || hash.includes('admin')) {
      this.isAdminRoute.set(true);
    } else if (path.includes('/merchant') || hash.includes('merchant')) {
      this.isMerchantRoute.set(true);
    }
  }

  fetchStores() {
    this.http.get<Store[]>('/api/v1/public/stores').subscribe({
      next: (data: Store[]) => this.stores.set(data || []),
      error: () => console.log('Mock stores fallback')
    });
  }

  fetchTariffs() {
    this.http.get<any>('/api/v1/public/settings/tariff').subscribe({
      next: (data: any) => {
        if (data && data.base_delivery_fee) {
          this.baseDeliveryFee.set(data.base_delivery_fee);
        }
      }
    });
  }

  loadStoreProducts(store: Store) {
    this.selectedStoreID.set(store.id);
    this.selectedStoreName.set(store.store_name);

    this.http.get<Product[]>(`/api/v1/public/products?store_id=${store.id}`).subscribe({
      next: (data: Product[]) => this.products.set(data || [])
    });
  }

  addToCart(product: Product) {
    this.cart.update((items: Product[]) => [...items, product]);
  }

  submitCheckout() {
    if (!this.customerName || !this.customerPhone) {
      alert('Mohon isi Nama dan Nomor WhatsApp!');
      return;
    }

    const payload = {
      store_id: this.selectedStoreID(),
      customer_name: this.customerName,
      customer_phone: this.customerPhone,
      delivery_address_text: this.customerAddress,
      items: this.cart().map((item: Product) => ({
        product_id: item.id,
        item_name: item.name,
        quantity: 1,
        price_per_item: item.price,
        notes: this.customNotes
      }))
    };

    this.http.post<any>('/api/v1/public/checkout', payload).subscribe({
      next: (res: any) => {
        if (this.paymentMethod === 'QRIS') {
          this.http.post<any>('/api/v1/payments/charge', {
            order_id: res.order_id,
            gross_amount: this.cartTotal(),
            payment_type: 'QRIS'
          }).subscribe({
            next: (payRes: any) => {
              this.qrisUrl.set(payRes.qr_code_url);
            }
          });
        } else {
          const waMsg = `Halo, saya ${this.customerName} mau pesan #${res.order_code}.\nAlamat: ${this.customerAddress}\nMohon diproses!`;
          window.location.href = `https://wa.me/${res.waha_bot_number}?text=${encodeURIComponent(waMsg)}`;
        }
      },
      error: () => alert('Gagal membuat pesanan!')
    });
  }
}
