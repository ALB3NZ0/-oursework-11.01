--
-- PostgreSQL database dump
--

-- Dumped from database version 17.4
-- Dumped by pg_dump version 17.4

-- Started on 2025-10-20 01:33:01

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

ALTER TABLE ONLY public.reviews DROP CONSTRAINT reviews_iduser_fkey;
ALTER TABLE ONLY public.reviews DROP CONSTRAINT reviews_idproduct_fkey;
ALTER TABLE ONLY public.reports DROP CONSTRAINT reports_iduser_fkey;
ALTER TABLE ONLY public.productsizes DROP CONSTRAINT productsizes_idproduct_fkey;
ALTER TABLE ONLY public.products DROP CONSTRAINT products_idcategory_fkey;
ALTER TABLE ONLY public.products DROP CONSTRAINT products_idbrand_fkey;
ALTER TABLE ONLY public.orders DROP CONSTRAINT orders_iduser_fkey;
ALTER TABLE ONLY public.orderproducts DROP CONSTRAINT orderproducts_idproductsize_fkey;
ALTER TABLE ONLY public.orderproducts DROP CONSTRAINT orderproducts_idorder_fkey;
ALTER TABLE ONLY public.logs DROP CONSTRAINT logs_iduser_fkey;
ALTER TABLE ONLY public.favorites DROP CONSTRAINT favorites_iduser_fkey;
ALTER TABLE ONLY public.favorites DROP CONSTRAINT favorites_idproductsize_fkey;
ALTER TABLE ONLY public.basket DROP CONSTRAINT basket_iduser_fkey;
ALTER TABLE ONLY public.basket DROP CONSTRAINT basket_idproductsize_fkey;
DROP TRIGGER trg_decrease_stock ON public.orderproducts;
DROP TRIGGER trg_check_product_quantity ON public.productsizes;
DROP TRIGGER trg_add_default_sizes ON public.products;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_pkey;
ALTER TABLE ONLY public.users DROP CONSTRAINT users_email_key;
ALTER TABLE ONLY public.reviews DROP CONSTRAINT reviews_pkey;
ALTER TABLE ONLY public.reports DROP CONSTRAINT reports_pkey;
ALTER TABLE ONLY public.productsizes DROP CONSTRAINT productsizes_pkey;
ALTER TABLE ONLY public.products DROP CONSTRAINT products_pkey;
ALTER TABLE ONLY public.orders DROP CONSTRAINT orders_pkey;
ALTER TABLE ONLY public.orderproducts DROP CONSTRAINT orderproducts_pkey;
ALTER TABLE ONLY public.logs DROP CONSTRAINT logs_pkey;
ALTER TABLE ONLY public.favorites DROP CONSTRAINT favorites_pkey;
ALTER TABLE ONLY public.categories DROP CONSTRAINT categories_pkey;
ALTER TABLE ONLY public.brands DROP CONSTRAINT brands_pkey;
ALTER TABLE ONLY public.basket DROP CONSTRAINT basket_pkey;
ALTER TABLE public.users ALTER COLUMN iduser DROP DEFAULT;
ALTER TABLE public.reviews ALTER COLUMN idreview DROP DEFAULT;
ALTER TABLE public.reports ALTER COLUMN idreport DROP DEFAULT;
ALTER TABLE public.productsizes ALTER COLUMN idproductsize DROP DEFAULT;
ALTER TABLE public.products ALTER COLUMN idproduct DROP DEFAULT;
ALTER TABLE public.orders ALTER COLUMN idorder DROP DEFAULT;
ALTER TABLE public.orderproducts ALTER COLUMN idorderproduct DROP DEFAULT;
ALTER TABLE public.logs ALTER COLUMN idlog DROP DEFAULT;
ALTER TABLE public.favorites ALTER COLUMN idfavorites DROP DEFAULT;
ALTER TABLE public.categories ALTER COLUMN idcategory DROP DEFAULT;
ALTER TABLE public.brands ALTER COLUMN idbrand DROP DEFAULT;
ALTER TABLE public.basket ALTER COLUMN idbasket DROP DEFAULT;
DROP SEQUENCE public.users_iduser_seq;
DROP VIEW public.topproducts;
DROP VIEW public.topcustomers;
DROP TABLE public.users;
DROP SEQUENCE public.reviews_idreview_seq;
DROP SEQUENCE public.reports_idreport_seq;
DROP TABLE public.reports;
DROP SEQUENCE public.productsizes_idproductsize_seq;
DROP SEQUENCE public.products_idproduct_seq;
DROP VIEW public.productratings;
DROP TABLE public.reviews;
DROP SEQUENCE public.orders_idorder_seq;
DROP SEQUENCE public.orderproducts_idorderproduct_seq;
DROP SEQUENCE public.logs_idlog_seq;
DROP TABLE public.logs;
DROP SEQUENCE public.favorites_idfavorites_seq;
DROP TABLE public.favorites;
DROP VIEW public.dailysales;
DROP TABLE public.productsizes;
DROP TABLE public.products;
DROP TABLE public.orders;
DROP TABLE public.orderproducts;
DROP SEQUENCE public.categories_idcategory_seq;
DROP TABLE public.categories;
DROP SEQUENCE public.brands_idbrand_seq;
DROP TABLE public.brands;
DROP SEQUENCE public.basket_idbasket_seq;
DROP TABLE public.basket;
DROP FUNCTION public.getusertotalspent(user_id integer);
DROP FUNCTION public.gettoprevenueproducts(start_date date, end_date date, limit_count integer);
DROP FUNCTION public.getrevenuebycategory(start_date date, end_date date);
DROP PROCEDURE public.generatetopcustomersreport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer);
DROP PROCEDURE public.generatesalesreport(IN start_date date, IN end_date date, IN report_name character varying, IN report_type character varying, IN user_id integer);
DROP PROCEDURE public.generatecategoryrevenuereport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer);
DROP FUNCTION public.decreasestock();
DROP FUNCTION public.check_product_quantity();
DROP FUNCTION public.adddefaultsizes();
--
-- TOC entry 264 (class 1255 OID 25267)
-- Name: adddefaultsizes(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.adddefaultsizes() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE
    s INT;
BEGIN
    FOR s IN 36..45 LOOP
        INSERT INTO ProductSizes(IdProduct, Size, Quantity)
        VALUES (NEW.IdProduct, s, 0);
    END LOOP;
    RETURN NEW;
END;
$$;


--
-- TOC entry 245 (class 1255 OID 33249)
-- Name: check_product_quantity(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.check_product_quantity() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    IF NEW.Quantity < 0 THEN
        RAISE EXCEPTION 'Количество товара не может быть отрицательным (IdProductSize=%)', NEW.IdProductSize;
    END IF;
    RETURN NEW;
END;
$$;


--
-- TOC entry 247 (class 1255 OID 25251)
-- Name: decreasestock(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.decreasestock() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    UPDATE ProductSizes
    SET Quantity = Quantity - NEW.Quantity
    WHERE IdProductSize = NEW.IdProductSize;
    RETURN NEW;
END;
$$;


--
-- TOC entry 263 (class 1255 OID 25265)
-- Name: generatecategoryrevenuereport(date, date, character varying, integer); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.generatecategoryrevenuereport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
    SELECT report_name, 
           'CategoryRevenue',
           string_agg(c.CategoryName || ': ' || SUM(op.Quantity * p.Price), ', '),
           user_id
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    JOIN Categories c ON p.IdCategory = c.IdCategory
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY user_id;
END;
$$;


--
-- TOC entry 261 (class 1255 OID 25262)
-- Name: generatesalesreport(date, date, character varying, character varying, integer); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.generatesalesreport(IN start_date date, IN end_date date, IN report_name character varying, IN report_type character varying, IN user_id integer)
    LANGUAGE plpgsql
    AS $$
DECLARE
    total_sales DECIMAL;
    total_orders INT;
BEGIN
    IF report_type = 'Sales' THEN
        SELECT COALESCE(SUM(op.Quantity * p.Price),0)
        INTO total_sales
        FROM Orders o
        JOIN OrderProducts op ON o.IdOrder = op.IdOrder
        JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
        JOIN Products p ON ps.IdProduct = p.IdProduct
        WHERE o.OrderDate BETWEEN start_date AND end_date;

        INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
        VALUES (report_name, report_type, 'Total Sales: ' || total_sales, user_id);

    ELSIF report_type = 'Orders' THEN
        SELECT COUNT(*)
        INTO total_orders
        FROM Orders
        WHERE OrderDate BETWEEN start_date AND end_date;

        INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
        VALUES (report_name, report_type, 'Total Orders: ' || total_orders, user_id);

    END IF;
END;
$$;


--
-- TOC entry 262 (class 1255 OID 25266)
-- Name: generatetopcustomersreport(date, date, character varying, integer); Type: PROCEDURE; Schema: public; Owner: -
--

CREATE PROCEDURE public.generatetopcustomersreport(IN start_date date, IN end_date date, IN report_name character varying, IN user_id integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    INSERT INTO Reports(ReportName, ReportType, ReportData, IdUser)
    SELECT report_name,
           'TopCustomers',
           string_agg(u.FullName || ': ' || SUM(op.Quantity * p.Price), ', '),
           user_id
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    JOIN Users u ON o.IdUser = u.IdUser
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY user_id
    ORDER BY SUM(op.Quantity * p.Price) DESC
    LIMIT 10;
END;
$$;


--
-- TOC entry 260 (class 1255 OID 25264)
-- Name: getrevenuebycategory(date, date); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.getrevenuebycategory(start_date date, end_date date) RETURNS TABLE(categoryname character varying, totalrevenue numeric, totalquantity integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT c.CategoryName,
           SUM(op.Quantity * p.Price) AS TotalRevenue,
           SUM(op.Quantity) AS TotalQuantity
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    JOIN Categories c ON p.IdCategory = c.IdCategory
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY c.CategoryName
    ORDER BY TotalRevenue DESC;
END;
$$;


--
-- TOC entry 259 (class 1255 OID 25263)
-- Name: gettoprevenueproducts(date, date, integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.gettoprevenueproducts(start_date date, end_date date, limit_count integer) RETURNS TABLE(productname character varying, totalrevenue numeric, totalquantity integer)
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN QUERY
    SELECT p.Name,
           SUM(op.Quantity * p.Price) AS TotalRevenue,
           SUM(op.Quantity) AS TotalQuantity
    FROM Orders o
    JOIN OrderProducts op ON o.IdOrder = op.IdOrder
    JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
    JOIN Products p ON ps.IdProduct = p.IdProduct
    WHERE o.OrderDate BETWEEN start_date AND end_date
    GROUP BY p.Name
    ORDER BY TotalRevenue DESC
    LIMIT limit_count;
END;
$$;


--
-- TOC entry 246 (class 1255 OID 25250)
-- Name: getusertotalspent(integer); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.getusertotalspent(user_id integer) RETURNS numeric
    LANGUAGE plpgsql
    AS $$
BEGIN
    RETURN (
        SELECT COALESCE(SUM(op.Quantity * p.Price), 0)
        FROM Orders o
        JOIN OrderProducts op ON o.IdOrder = op.IdOrder
        JOIN ProductSizes ps ON op.IdProductSize = ps.IdProductSize
        JOIN Products p ON ps.IdProduct = p.IdProduct
        WHERE o.IdUser = user_id
    );
END;
$$;


SET default_table_access_method = heap;

--
-- TOC entry 228 (class 1259 OID 25115)
-- Name: basket; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.basket (
    idbasket integer NOT NULL,
    iduser integer NOT NULL,
    idproductsize integer NOT NULL,
    quantity integer NOT NULL
);


--
-- TOC entry 227 (class 1259 OID 25114)
-- Name: basket_idbasket_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.basket_idbasket_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5060 (class 0 OID 0)
-- Dependencies: 227
-- Name: basket_idbasket_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.basket_idbasket_seq OWNED BY public.basket.idbasket;


--
-- TOC entry 220 (class 1259 OID 25069)
-- Name: brands; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.brands (
    idbrand integer NOT NULL,
    brandname character varying(100) NOT NULL
);


--
-- TOC entry 219 (class 1259 OID 25068)
-- Name: brands_idbrand_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.brands_idbrand_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5061 (class 0 OID 0)
-- Dependencies: 219
-- Name: brands_idbrand_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.brands_idbrand_seq OWNED BY public.brands.idbrand;


--
-- TOC entry 222 (class 1259 OID 25077)
-- Name: categories; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.categories (
    idcategory integer NOT NULL,
    categoryname character varying(100) NOT NULL
);


--
-- TOC entry 221 (class 1259 OID 25076)
-- Name: categories_idcategory_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.categories_idcategory_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5062 (class 0 OID 0)
-- Dependencies: 221
-- Name: categories_idcategory_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.categories_idcategory_seq OWNED BY public.categories.idcategory;


--
-- TOC entry 236 (class 1259 OID 25183)
-- Name: orderproducts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orderproducts (
    idorderproduct integer NOT NULL,
    idorder integer,
    idproductsize integer,
    quantity integer NOT NULL
);


--
-- TOC entry 234 (class 1259 OID 25170)
-- Name: orders; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.orders (
    idorder integer NOT NULL,
    iduser integer,
    orderdate timestamp without time zone DEFAULT now()
);


--
-- TOC entry 224 (class 1259 OID 25084)
-- Name: products; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.products (
    idproduct integer NOT NULL,
    name character varying(100) NOT NULL,
    imageurl character varying(255),
    price numeric NOT NULL,
    idbrand integer NOT NULL,
    idcategory integer NOT NULL
);


--
-- TOC entry 226 (class 1259 OID 25103)
-- Name: productsizes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.productsizes (
    idproductsize integer NOT NULL,
    idproduct integer NOT NULL,
    size integer NOT NULL,
    quantity integer NOT NULL
);


--
-- TOC entry 242 (class 1259 OID 25235)
-- Name: dailysales; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.dailysales AS
 SELECT date(o.orderdate) AS day,
    sum(((op.quantity)::numeric * p.price)) AS revenue
   FROM (((public.orders o
     JOIN public.orderproducts op ON ((o.idorder = op.idorder)))
     JOIN public.productsizes ps ON ((op.idproductsize = ps.idproductsize)))
     JOIN public.products p ON ((ps.idproduct = p.idproduct)))
  GROUP BY (date(o.orderdate))
  ORDER BY (date(o.orderdate)) DESC;


--
-- TOC entry 230 (class 1259 OID 25132)
-- Name: favorites; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.favorites (
    idfavorites integer NOT NULL,
    iduser integer NOT NULL,
    idproductsize integer NOT NULL
);


--
-- TOC entry 229 (class 1259 OID 25131)
-- Name: favorites_idfavorites_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.favorites_idfavorites_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5063 (class 0 OID 0)
-- Dependencies: 229
-- Name: favorites_idfavorites_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.favorites_idfavorites_seq OWNED BY public.favorites.idfavorites;


--
-- TOC entry 240 (class 1259 OID 25215)
-- Name: logs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.logs (
    idlog integer NOT NULL,
    iduser integer NOT NULL,
    action character varying(100) NOT NULL,
    entity character varying(50),
    entityid integer,
    details text,
    createdat timestamp without time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 239 (class 1259 OID 25214)
-- Name: logs_idlog_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.logs_idlog_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5064 (class 0 OID 0)
-- Dependencies: 239
-- Name: logs_idlog_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.logs_idlog_seq OWNED BY public.logs.idlog;


--
-- TOC entry 235 (class 1259 OID 25182)
-- Name: orderproducts_idorderproduct_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orderproducts_idorderproduct_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5065 (class 0 OID 0)
-- Dependencies: 235
-- Name: orderproducts_idorderproduct_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.orderproducts_idorderproduct_seq OWNED BY public.orderproducts.idorderproduct;


--
-- TOC entry 233 (class 1259 OID 25169)
-- Name: orders_idorder_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.orders_idorder_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5066 (class 0 OID 0)
-- Dependencies: 233
-- Name: orders_idorder_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.orders_idorder_seq OWNED BY public.orders.idorder;


--
-- TOC entry 232 (class 1259 OID 25149)
-- Name: reviews; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reviews (
    idreview integer NOT NULL,
    idproduct integer NOT NULL,
    rating integer NOT NULL,
    comment text,
    reviewdate timestamp without time zone DEFAULT now() NOT NULL,
    iduser integer,
    CONSTRAINT reviews_rating_check CHECK (((rating >= 1) AND (rating <= 5)))
);


--
-- TOC entry 243 (class 1259 OID 25240)
-- Name: productratings; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.productratings AS
 SELECT p.name,
    round(avg(r.rating), 2) AS avgrating,
    count(r.idreview) AS reviewcount
   FROM (public.products p
     LEFT JOIN public.reviews r ON ((p.idproduct = r.idproduct)))
  GROUP BY p.name;


--
-- TOC entry 223 (class 1259 OID 25083)
-- Name: products_idproduct_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.products_idproduct_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5067 (class 0 OID 0)
-- Dependencies: 223
-- Name: products_idproduct_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.products_idproduct_seq OWNED BY public.products.idproduct;


--
-- TOC entry 225 (class 1259 OID 25102)
-- Name: productsizes_idproductsize_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.productsizes_idproductsize_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5068 (class 0 OID 0)
-- Dependencies: 225
-- Name: productsizes_idproductsize_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.productsizes_idproductsize_seq OWNED BY public.productsizes.idproductsize;


--
-- TOC entry 238 (class 1259 OID 25200)
-- Name: reports; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.reports (
    idreport integer NOT NULL,
    reportname character varying(255) NOT NULL,
    reporttype character varying(50) NOT NULL,
    reportdata text NOT NULL,
    iduser integer NOT NULL,
    createdat timestamp without time zone DEFAULT now() NOT NULL
);


--
-- TOC entry 237 (class 1259 OID 25199)
-- Name: reports_idreport_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reports_idreport_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5069 (class 0 OID 0)
-- Dependencies: 237
-- Name: reports_idreport_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reports_idreport_seq OWNED BY public.reports.idreport;


--
-- TOC entry 231 (class 1259 OID 25148)
-- Name: reviews_idreview_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.reviews_idreview_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5070 (class 0 OID 0)
-- Dependencies: 231
-- Name: reviews_idreview_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.reviews_idreview_seq OWNED BY public.reviews.idreview;


--
-- TOC entry 218 (class 1259 OID 25058)
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    iduser integer NOT NULL,
    fullname character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    passwordhash character varying(255) NOT NULL,
    roleid integer NOT NULL
);


--
-- TOC entry 244 (class 1259 OID 25244)
-- Name: topcustomers; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.topcustomers AS
 SELECT u.fullname,
    u.email,
    sum(((op.quantity)::numeric * p.price)) AS totalspent
   FROM ((((public.orders o
     JOIN public.orderproducts op ON ((o.idorder = op.idorder)))
     JOIN public.productsizes ps ON ((op.idproductsize = ps.idproductsize)))
     JOIN public.products p ON ((ps.idproduct = p.idproduct)))
     JOIN public.users u ON ((o.iduser = u.iduser)))
  GROUP BY u.iduser, u.fullname, u.email
  ORDER BY (sum(((op.quantity)::numeric * p.price))) DESC
 LIMIT 10;


--
-- TOC entry 241 (class 1259 OID 25230)
-- Name: topproducts; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.topproducts AS
 SELECT p.name,
    sum(op.quantity) AS totalsold
   FROM ((public.orderproducts op
     JOIN public.productsizes ps ON ((op.idproductsize = ps.idproductsize)))
     JOIN public.products p ON ((ps.idproduct = p.idproduct)))
  GROUP BY p.name
  ORDER BY (sum(op.quantity)) DESC
 LIMIT 10;


--
-- TOC entry 217 (class 1259 OID 25057)
-- Name: users_iduser_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_iduser_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- TOC entry 5071 (class 0 OID 0)
-- Dependencies: 217
-- Name: users_iduser_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_iduser_seq OWNED BY public.users.iduser;


--
-- TOC entry 4827 (class 2604 OID 25118)
-- Name: basket idbasket; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket ALTER COLUMN idbasket SET DEFAULT nextval('public.basket_idbasket_seq'::regclass);


--
-- TOC entry 4823 (class 2604 OID 25072)
-- Name: brands idbrand; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.brands ALTER COLUMN idbrand SET DEFAULT nextval('public.brands_idbrand_seq'::regclass);


--
-- TOC entry 4824 (class 2604 OID 25080)
-- Name: categories idcategory; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories ALTER COLUMN idcategory SET DEFAULT nextval('public.categories_idcategory_seq'::regclass);


--
-- TOC entry 4828 (class 2604 OID 25135)
-- Name: favorites idfavorites; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites ALTER COLUMN idfavorites SET DEFAULT nextval('public.favorites_idfavorites_seq'::regclass);


--
-- TOC entry 4836 (class 2604 OID 25218)
-- Name: logs idlog; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs ALTER COLUMN idlog SET DEFAULT nextval('public.logs_idlog_seq'::regclass);


--
-- TOC entry 4833 (class 2604 OID 25186)
-- Name: orderproducts idorderproduct; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts ALTER COLUMN idorderproduct SET DEFAULT nextval('public.orderproducts_idorderproduct_seq'::regclass);


--
-- TOC entry 4831 (class 2604 OID 25173)
-- Name: orders idorder; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders ALTER COLUMN idorder SET DEFAULT nextval('public.orders_idorder_seq'::regclass);


--
-- TOC entry 4825 (class 2604 OID 25087)
-- Name: products idproduct; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products ALTER COLUMN idproduct SET DEFAULT nextval('public.products_idproduct_seq'::regclass);


--
-- TOC entry 4826 (class 2604 OID 25106)
-- Name: productsizes idproductsize; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.productsizes ALTER COLUMN idproductsize SET DEFAULT nextval('public.productsizes_idproductsize_seq'::regclass);


--
-- TOC entry 4834 (class 2604 OID 25203)
-- Name: reports idreport; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports ALTER COLUMN idreport SET DEFAULT nextval('public.reports_idreport_seq'::regclass);


--
-- TOC entry 4829 (class 2604 OID 25152)
-- Name: reviews idreview; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews ALTER COLUMN idreview SET DEFAULT nextval('public.reviews_idreview_seq'::regclass);


--
-- TOC entry 4822 (class 2604 OID 25061)
-- Name: users iduser; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN iduser SET DEFAULT nextval('public.users_iduser_seq'::regclass);


--
-- TOC entry 5042 (class 0 OID 25115)
-- Dependencies: 228
-- Data for Name: basket; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.basket (idbasket, iduser, idproductsize, quantity) FROM stdin;
1	3	1	1
2	3	2	2
3	4	5	1
\.


--
-- TOC entry 5034 (class 0 OID 25069)
-- Dependencies: 220
-- Data for Name: brands; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.brands (idbrand, brandname) FROM stdin;
1	Nike
2	Adidas
3	Puma
4	Reebok
\.


--
-- TOC entry 5036 (class 0 OID 25077)
-- Dependencies: 222
-- Data for Name: categories; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.categories (idcategory, categoryname) FROM stdin;
1	Кроссовки
2	Ботинки
3	Кеды
\.


--
-- TOC entry 5044 (class 0 OID 25132)
-- Dependencies: 230
-- Data for Name: favorites; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.favorites (idfavorites, iduser, idproductsize) FROM stdin;
1	3	1
2	4	2
3	5	3
\.


--
-- TOC entry 5054 (class 0 OID 25215)
-- Dependencies: 240
-- Data for Name: logs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.logs (idlog, iduser, action, entity, entityid, details, createdat) FROM stdin;
14	18	CREATE	user/reports	\N	user CREATE reports	2025-10-17 12:29:49.957029
15	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 00:45:55.423885
16	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:00:17.53351
17	16	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:01:31.402638
18	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:02:05.711891
19	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:10:55.391486
20	17	CREATE	backup	0	Создан бэкап базы данных: shoes_store_backup_20251020_011348.sql (0.04 MB)	2025-10-20 01:13:48.812348
21	17	CREATE	admin/backup	\N	admin CREATE backup	2025-10-20 01:13:48.184352
\.


--
-- TOC entry 5050 (class 0 OID 25183)
-- Dependencies: 236
-- Data for Name: orderproducts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.orderproducts (idorderproduct, idorder, idproductsize, quantity) FROM stdin;
4	1	1	1
5	1	2	1
6	2	5	1
\.


--
-- TOC entry 5048 (class 0 OID 25170)
-- Dependencies: 234
-- Data for Name: orders; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.orders (idorder, iduser, orderdate) FROM stdin;
1	3	2025-09-28 16:27:41.528981
2	4	2025-09-28 16:27:41.528981
\.


--
-- TOC entry 5038 (class 0 OID 25084)
-- Dependencies: 224
-- Data for Name: products; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.products (idproduct, name, imageurl, price, idbrand, idcategory) FROM stdin;
1	Nike Air Max	nike_airmax.jpg	150.00	1	1
2	Adidas Ultraboost	adidas_ultraboost.jpg	180.00	2	1
3	Puma Runner	puma_runner.jpg	120.00	3	3
4	Reebok Classic	reebok_classic.jpg	130.00	4	3
5	Nike Air Force 1	nike_af1.jpg	160.00	1	1
7	Adidas Samba	https://clck.ru/3PiDUw	10000	2	3
\.


--
-- TOC entry 5040 (class 0 OID 25103)
-- Dependencies: 226
-- Data for Name: productsizes; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.productsizes (idproductsize, idproduct, size, quantity) FROM stdin;
61	7	36	0
62	7	37	0
63	7	38	0
64	7	39	0
65	7	40	0
66	7	41	0
67	7	42	0
68	7	43	0
69	7	44	0
70	7	45	0
3	1	38	100
4	1	39	100
6	1	41	100
7	1	42	100
8	1	43	100
9	1	44	100
10	1	45	100
11	2	36	100
12	2	37	100
13	2	38	100
14	2	39	100
15	2	40	100
16	2	41	100
17	2	42	100
18	2	43	100
19	2	44	100
20	2	45	100
21	3	36	100
22	3	37	100
23	3	38	100
24	3	39	100
25	3	40	100
26	3	41	100
27	3	42	100
28	3	43	100
29	3	44	100
30	3	45	100
31	4	36	100
32	4	37	100
33	4	38	100
34	4	39	100
35	4	40	100
36	4	41	100
37	4	42	100
38	4	43	100
39	4	44	100
40	4	45	100
41	5	36	100
42	5	37	100
43	5	38	100
44	5	39	100
45	5	40	100
46	5	41	100
47	5	42	100
48	5	43	100
49	5	44	100
50	5	45	100
1	1	36	100
2	1	37	100
5	1	40	100
\.


--
-- TOC entry 5052 (class 0 OID 25200)
-- Dependencies: 238
-- Data for Name: reports; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reports (idreport, reportname, reporttype, reportdata, iduser, createdat) FROM stdin;
\.


--
-- TOC entry 5046 (class 0 OID 25149)
-- Dependencies: 232
-- Data for Name: reviews; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.reviews (idreview, idproduct, rating, comment, reviewdate, iduser) FROM stdin;
1	1	5	Отличные кроссовки!	2025-09-28 16:34:46.024037	3
2	2	4	Очень удобные	2025-09-28 16:34:46.024037	4
3	3	3	Нормально для пробежки	2025-09-28 16:34:46.024037	5
4	5	5	Классические Air Force 1, рекомендую	2025-09-28 16:34:46.024037	3
\.


--
-- TOC entry 5032 (class 0 OID 25058)
-- Dependencies: 218
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (iduser, fullname, email, passwordhash, roleid) FROM stdin;
1	Иван Иванов	admin@example.com	hashedpassword1	1
2	Мария Петрова	manager@example.com	hashedpassword2	2
3	Алексей Сидоров	user1@example.com	hashedpassword3	3
4	Ольга Смирнова	user2@example.com	hashedpassword4	3
5	Дмитрий Кузнецов	user3@example.com	hashedpassword5	3
10	Rusik	rus@rus.ru	passwordhash6	3
14	zxczxc	zxc@zxc.zxc	$2a$10$iKc0uuK9lOCjICWmIzT6iu14OEBqRu57RRC7jLDakQEIgHbQKXnW.	3
13	dsa	dsa@dsa.dsa	$2a$10$BNQ0BiIN0YJn.VUtvbqeIOo6tdVtZCgL5cCM2zO4JYNCLBtffvmHe	3
15	op	op@op.op	$2a$10$751kb2914A378xROAOcxKuE4dA5ewUvDS6W8761DNI16zoBKYgUhG	3
16	rus	rus@gmail.com	$2a$10$0Q6MYwO3PQ9m6dgfUPBQRODtMqS66EMBssd0ugdZhmRxEh8UZq6b2	3
17	admin	admin@gmail.com	$2a$10$IRhFP0CQtnDtNsx4um1Ov.o2CQm.ba.xLgLar58j6N30LysTf42MO	1
18	manager	manager@gmail.com	$2a$10$cgiVJg1b9q6fVNSZsXVbduAAd0Yy/XcBc.RVNCKd3lPaoLTbHTB.6	2
\.


--
-- TOC entry 5072 (class 0 OID 0)
-- Dependencies: 227
-- Name: basket_idbasket_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.basket_idbasket_seq', 3, true);


--
-- TOC entry 5073 (class 0 OID 0)
-- Dependencies: 219
-- Name: brands_idbrand_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.brands_idbrand_seq', 6, true);


--
-- TOC entry 5074 (class 0 OID 0)
-- Dependencies: 221
-- Name: categories_idcategory_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.categories_idcategory_seq', 4, true);


--
-- TOC entry 5075 (class 0 OID 0)
-- Dependencies: 229
-- Name: favorites_idfavorites_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.favorites_idfavorites_seq', 3, true);


--
-- TOC entry 5076 (class 0 OID 0)
-- Dependencies: 239
-- Name: logs_idlog_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.logs_idlog_seq', 21, true);


--
-- TOC entry 5077 (class 0 OID 0)
-- Dependencies: 235
-- Name: orderproducts_idorderproduct_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.orderproducts_idorderproduct_seq', 6, true);


--
-- TOC entry 5078 (class 0 OID 0)
-- Dependencies: 233
-- Name: orders_idorder_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.orders_idorder_seq', 2, true);


--
-- TOC entry 5079 (class 0 OID 0)
-- Dependencies: 223
-- Name: products_idproduct_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.products_idproduct_seq', 7, true);


--
-- TOC entry 5080 (class 0 OID 0)
-- Dependencies: 225
-- Name: productsizes_idproductsize_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.productsizes_idproductsize_seq', 70, true);


--
-- TOC entry 5081 (class 0 OID 0)
-- Dependencies: 237
-- Name: reports_idreport_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.reports_idreport_seq', 1, false);


--
-- TOC entry 5082 (class 0 OID 0)
-- Dependencies: 231
-- Name: reviews_idreview_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.reviews_idreview_seq', 6, true);


--
-- TOC entry 5083 (class 0 OID 0)
-- Dependencies: 217
-- Name: users_iduser_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.users_iduser_seq', 18, true);


--
-- TOC entry 4852 (class 2606 OID 25120)
-- Name: basket basket_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket
    ADD CONSTRAINT basket_pkey PRIMARY KEY (idbasket);


--
-- TOC entry 4844 (class 2606 OID 25074)
-- Name: brands brands_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.brands
    ADD CONSTRAINT brands_pkey PRIMARY KEY (idbrand);


--
-- TOC entry 4846 (class 2606 OID 25082)
-- Name: categories categories_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.categories
    ADD CONSTRAINT categories_pkey PRIMARY KEY (idcategory);


--
-- TOC entry 4854 (class 2606 OID 25137)
-- Name: favorites favorites_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_pkey PRIMARY KEY (idfavorites);


--
-- TOC entry 4864 (class 2606 OID 25223)
-- Name: logs logs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_pkey PRIMARY KEY (idlog);


--
-- TOC entry 4860 (class 2606 OID 25188)
-- Name: orderproducts orderproducts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts
    ADD CONSTRAINT orderproducts_pkey PRIMARY KEY (idorderproduct);


--
-- TOC entry 4858 (class 2606 OID 25176)
-- Name: orders orders_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_pkey PRIMARY KEY (idorder);


--
-- TOC entry 4848 (class 2606 OID 25091)
-- Name: products products_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_pkey PRIMARY KEY (idproduct);


--
-- TOC entry 4850 (class 2606 OID 25108)
-- Name: productsizes productsizes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.productsizes
    ADD CONSTRAINT productsizes_pkey PRIMARY KEY (idproductsize);


--
-- TOC entry 4862 (class 2606 OID 25208)
-- Name: reports reports_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_pkey PRIMARY KEY (idreport);


--
-- TOC entry 4856 (class 2606 OID 25158)
-- Name: reviews reviews_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_pkey PRIMARY KEY (idreview);


--
-- TOC entry 4840 (class 2606 OID 25067)
-- Name: users users_email_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);


--
-- TOC entry 4842 (class 2606 OID 25065)
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (iduser);


--
-- TOC entry 4879 (class 2620 OID 25268)
-- Name: products trg_add_default_sizes; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_add_default_sizes AFTER INSERT ON public.products FOR EACH ROW EXECUTE FUNCTION public.adddefaultsizes();


--
-- TOC entry 4880 (class 2620 OID 33250)
-- Name: productsizes trg_check_product_quantity; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_check_product_quantity BEFORE INSERT OR UPDATE ON public.productsizes FOR EACH ROW EXECUTE FUNCTION public.check_product_quantity();


--
-- TOC entry 4881 (class 2620 OID 25252)
-- Name: orderproducts trg_decrease_stock; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_decrease_stock AFTER INSERT ON public.orderproducts FOR EACH ROW EXECUTE FUNCTION public.decreasestock();


--
-- TOC entry 4868 (class 2606 OID 25126)
-- Name: basket basket_idproductsize_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket
    ADD CONSTRAINT basket_idproductsize_fkey FOREIGN KEY (idproductsize) REFERENCES public.productsizes(idproductsize);


--
-- TOC entry 4869 (class 2606 OID 25121)
-- Name: basket basket_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.basket
    ADD CONSTRAINT basket_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4870 (class 2606 OID 25143)
-- Name: favorites favorites_idproductsize_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_idproductsize_fkey FOREIGN KEY (idproductsize) REFERENCES public.productsizes(idproductsize);


--
-- TOC entry 4871 (class 2606 OID 25138)
-- Name: favorites favorites_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.favorites
    ADD CONSTRAINT favorites_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4878 (class 2606 OID 25224)
-- Name: logs logs_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.logs
    ADD CONSTRAINT logs_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4875 (class 2606 OID 25189)
-- Name: orderproducts orderproducts_idorder_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts
    ADD CONSTRAINT orderproducts_idorder_fkey FOREIGN KEY (idorder) REFERENCES public.orders(idorder);


--
-- TOC entry 4876 (class 2606 OID 25194)
-- Name: orderproducts orderproducts_idproductsize_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orderproducts
    ADD CONSTRAINT orderproducts_idproductsize_fkey FOREIGN KEY (idproductsize) REFERENCES public.productsizes(idproductsize);


--
-- TOC entry 4874 (class 2606 OID 25177)
-- Name: orders orders_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.orders
    ADD CONSTRAINT orders_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4865 (class 2606 OID 25092)
-- Name: products products_idbrand_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_idbrand_fkey FOREIGN KEY (idbrand) REFERENCES public.brands(idbrand);


--
-- TOC entry 4866 (class 2606 OID 25097)
-- Name: products products_idcategory_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.products
    ADD CONSTRAINT products_idcategory_fkey FOREIGN KEY (idcategory) REFERENCES public.categories(idcategory);


--
-- TOC entry 4867 (class 2606 OID 25109)
-- Name: productsizes productsizes_idproduct_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.productsizes
    ADD CONSTRAINT productsizes_idproduct_fkey FOREIGN KEY (idproduct) REFERENCES public.products(idproduct);


--
-- TOC entry 4877 (class 2606 OID 25209)
-- Name: reports reports_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reports
    ADD CONSTRAINT reports_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


--
-- TOC entry 4872 (class 2606 OID 25159)
-- Name: reviews reviews_idproduct_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_idproduct_fkey FOREIGN KEY (idproduct) REFERENCES public.products(idproduct);


--
-- TOC entry 4873 (class 2606 OID 25164)
-- Name: reviews reviews_iduser_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.reviews
    ADD CONSTRAINT reviews_iduser_fkey FOREIGN KEY (iduser) REFERENCES public.users(iduser);


-- Completed on 2025-10-20 01:33:02

--
-- PostgreSQL database dump complete
--

