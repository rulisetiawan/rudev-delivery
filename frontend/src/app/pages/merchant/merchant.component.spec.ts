import { ComponentFixture, TestBed } from '@angular/core/testing';
import { HttpClientTestingModule } from '@angular/common/http/testing';
import { MerchantDashboardComponent } from './merchant.component';

describe('MerchantDashboardComponent (Merchant Web Portal Test)', () => {
  let component: MerchantDashboardComponent;
  let fixture: ComponentFixture<MerchantDashboardComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MerchantDashboardComponent, HttpClientTestingModule]
    }).compileComponents();

    fixture = TestBed.createComponent(MerchantDashboardComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create the merchant dashboard component', () => {
    expect(component).toBeTruthy();
  });

  it('should initialize with an empty products signal array', () => {
    expect(component.products().length).toBe(0);
  });

  it('should open and close the add product modal', () => {
    expect(component.displayAddProductModal()).toBeFalse();
    component.displayAddProductModal.set(true);
    expect(component.displayAddProductModal()).toBeTrue();
  });
});
