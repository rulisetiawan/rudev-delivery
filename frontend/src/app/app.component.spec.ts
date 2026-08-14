import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { AppComponent } from './app.component';

describe('AppComponent (Customer PWA & Angular Signals Test)', () => {
  let component: AppComponent;
  let fixture: ComponentFixture<AppComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [AppComponent, HttpClientTestingModule]
    }).compileComponents();

    fixture = TestBed.createComponent(AppComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create the app component', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize with an empty cart and count 0', () => {
    expect(component.cartCount()).toBe(0);
    expect(component.cartSubtotal()).toBe(0);
  });

  it('should correctly update Signals cart count and total when adding items', () => {
    const mockProduct = {
      id: 101,
      store_id: 1,
      name: 'Nasi Goreng Spesial',
      price: 15000,
      category: 'Makanan',
      is_available: true,
      image_url: ''
    };

    component.addToCart(mockProduct);

    expect(component.cartCount()).toBe(1);
    expect(component.cartSubtotal()).toBe(15000);
    expect(component.cartTotal()).toBe(25000); // 15000 + 10000 base delivery fee
  });
});
